package agent

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/did"
	"go.bryk.io/pkg/did/resolver"
	"go.bryk.io/pkg/log"
	"go.bryk.io/pkg/net/rpc"
	"google.golang.org/grpc"
)

// Handler provides the required functionality for the DID method.
type Handler struct {
	methods     []string
	store       Storage
	log         log.Logger
	difficulty  uint
	algoNode    *algod.Client
	algoIndexer *indexer.Client
}

// HandlerOptions defines the settings available to adjust the
// operation of a handler instance.
type HandlerOptions struct {
	// Supported DID methods.
	Methods []string

	// PoW difficulty level required for valid transactions.
	Difficulty uint

	// Storage mechanism to be used for persistent state.
	Store Storage

	// Log sink.
	Logger log.Logger

	// Algorand node client.
	AlgoNode *algod.Client

	// Algorand indexer client.
	AlgoIndexer *indexer.Client
}

// NewHandler starts a new DID method handler instance.
func NewHandler(options HandlerOptions) (*Handler, error) {
	return &Handler{
		log:         options.Logger,
		store:       options.Store,
		methods:     options.Methods,
		difficulty:  options.Difficulty,
		algoNode:    options.AlgoNode,
		algoIndexer: options.AlgoIndexer,
	}, nil
}

// Close the instance and safely terminate any internal processing.
func (h *Handler) Close() error {
	h.log.Info("closing agent handler")
	return h.store.Close()
}

// Retrieve an existing DID instance based on its subject string.
func (h *Handler) Retrieve(req *protoV1.QueryRequest) (*did.Identifier, *did.ProofLD, error) {
	logFields := log.Fields{
		"method":  req.Method,
		"subject": req.Subject,
	}
	h.log.WithFields(logFields).Debug("retrieve request")

	// Verify method is supported
	if !h.isSupported(req.Method) {
		h.log.WithFields(logFields).Warning("non supported method")
		return nil, nil, errors.New(resolver.ErrMethodNotSupported)
	}

	// Retrieve document from storage
	id, proof, err := h.store.Get(req)
	if err != nil {
		h.log.WithFields(logFields).Warning(err.Error())
		return nil, nil, errors.New(resolver.ErrNotFound)
	}
	return id, proof, nil
}

// Process an incoming request ticket.
func (h *Handler) Process(req *protoV1.ProcessRequest) (string, error) {
	// Empty request
	if req == nil {
		return "", errors.New("empty request")
	}

	// Validate ticket
	if err := req.Ticket.Verify(h.difficulty); err != nil {
		h.log.WithFields(log.Fields{"error": err.Error()}).Error("invalid ticket")
		return "", err
	}

	// Load DID document and proof
	id, err := req.Ticket.GetDID()
	if err != nil {
		h.log.WithFields(log.Fields{"error": err.Error()}).Error("invalid DID contents")
		return "", err
	}
	proof, err := req.Ticket.GetProofLD()
	if err != nil {
		h.log.WithFields(log.Fields{"error": err.Error()}).Error("invalid DID proof")
		return "", err
	}

	// Verify method is supported
	if !h.isSupported(id.Method()) {
		h.log.WithFields(log.Fields{"method": id.Method()}).Warning("non supported method")
		return "", errors.New("non supported method")
	}

	// Update operations require another validation step using the original record
	isUpdate := h.store.Exists(id)
	if isUpdate {
		if err := req.Ticket.Verify(h.difficulty); err != nil {
			h.log.WithFields(log.Fields{"error": err.Error()}).Error("invalid ticket")
			return "", err
		}
	}

	// Store record
	h.log.WithFields(log.Fields{
		"subject": id.Subject(),
		"update":  isUpdate,
		"task":    req.Task,
	}).Debug("write operation")
	return h.store.Save(id, proof)
}

// AccountInformation returns details about the crypto account specified.
func (h *Handler) AccountInformation(ctx context.Context, req *protoV1.AccountInformationRequest) (*protoV1.AccountInformationResponse, error) { // nolint: lll
	ai, err := h.algoNode.AccountInformation(req.Address).Do(ctx)
	if err != nil {
		h.log.WithFields(log.Fields{
			"error":   err.Error(),
			"address": req.Address,
		}).Error("failed to get account information")
		return nil, err
	}
	_, ptList, err := h.algoNode.PendingTransactionsByAddress(req.Address).Do(ctx)
	if err != nil {
		h.log.WithFields(log.Fields{
			"error":   err.Error(),
			"address": req.Address,
		}).Error("failed to get pending transactions")
		return nil, err
	}
	resp := &protoV1.AccountInformationResponse{
		Status:              ai.Status,
		Balance:             ai.AmountWithoutPendingRewards,
		TotalRewards:        ai.Rewards,
		PendingRewards:      ai.PendingRewards,
		PendingTransactions: []*protoV1.AlgoTransaction{},
	}
	for _, pt := range ptList {
		resp.PendingTransactions = append(resp.PendingTransactions, &protoV1.AlgoTransaction{
			Amount:   uint64(pt.Txn.Amount),
			Receiver: pt.Txn.Receiver.String(),
			Note:     pt.Txn.Note,
		})
	}
	return resp, nil
}

// AccountActivity opens a channel to monitor near real-time account activity.
// The channel must be closed using the provided context when no longer needed.
func (h *Handler) AccountActivity(ctx context.Context, req *protoV1.AccountActivityRequest) (<-chan *protoV1.AccountActivityResponse, error) { // nolint: lll
	check := time.NewTicker(time.Duration(5) * time.Second)
	sink := make(chan *protoV1.AccountActivityResponse)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Close channel, monitor and processing routine
				close(sink)
				check.Stop()
				return
			case <-check.C:
				query := h.algoIndexer.LookupAccountTransactions(req.Address)
				resp, err := query.Do(ctx)
				if err != nil {
					h.log.WithFields(log.Fields{
						"error":   err.Error(),
						"address": req.Address,
					}).Error("failed to get account activity")
				} else {
					sink <- &protoV1.AccountActivityResponse{
						CurrentRound: resp.CurrentRound,
						NextToken:    resp.NextToken,
					}
				}
			}
		}
	}()
	return sink, nil
}

// TxParameters return the latest network parameters suggested for processing
// new transactions.
func (h *Handler) TxParameters(ctx context.Context) (*protoV1.TxParametersResponse, error) {
	params, err := h.algoNode.SuggestedParams().Do(ctx)
	if err != nil {
		h.log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get transaction parameters")
		return nil, err
	}
	data, err := json.Marshal(params)
	if err != nil {
		h.log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to encode transaction parameters")
		return nil, err
	}
	return &protoV1.TxParametersResponse{Params: data}, nil
}

// TxSubmit will send a signed raw transaction to the network for processing.
func (h *Handler) TxSubmit(ctx context.Context, req *protoV1.TxSubmitRequest) (*protoV1.TxSubmitResponse, error) {
	tid, err := h.algoNode.SendRawTransaction(req.Stx).Do(ctx)
	if err != nil {
		h.log.WithFields(log.Fields{
			"error": err.Error(),
			"tx":    base64.StdEncoding.EncodeToString(req.Stx),
		}).Error("failed to submit raw transaction")
		return nil, err
	}
	return &protoV1.TxSubmitResponse{Id: tid}, nil
}

// ServerSetup performs all initialization requirements for the
// handler instance to be exposed through the provided gRPC server.
func (h *Handler) ServerSetup(server *grpc.Server) {
	protoV1.RegisterAgentAPIServer(server, &rpcHandler{handler: h})
}

// GatewaySetup return the HTTP setup method to allow exposing the
// handler's functionality through an HTTP gateway.
func (h *Handler) GatewaySetup() rpc.GatewayRegisterFunc {
	return protoV1.RegisterAgentAPIHandler
}

// QueryResponseFilter provides custom encoding of HTTP query results.
func (h *Handler) QueryResponseFilter() rpc.GatewayInterceptor {
	return func(res http.ResponseWriter, req *http.Request) error {
		// Filter query requests
		if !strings.HasPrefix(req.URL.Path, "/v1/retrieve/") {
			return nil
		}
		seg := strings.Split(strings.TrimPrefix(req.URL.Path, "/v1/retrieve/"), "/")
		if len(seg) != 2 {
			return nil
		}

		// Submit query
		var (
			status   = http.StatusNotFound
			response []byte
		)
		rr := &protoV1.QueryRequest{
			Method:  seg[0],
			Subject: seg[1],
		}
		id, proof, err := h.Retrieve(rr)
		if err != nil {
			response, _ = json.MarshalIndent(map[string]string{"error": err.Error()}, "", "  ")
		} else {
			response, _ = json.MarshalIndent(map[string]interface{}{
				"document": id.Document(true),
				"proof":    proof,
				"metadata": id.GetMetadata(),
			}, "", "  ")
			status = http.StatusOK
			res.Header().Set("Etag", fmt.Sprintf("W/%x", sha256.Sum256(response)))
		}

		// Return result
		res.WriteHeader(status)
		_, _ = res.Write(response)
		return errors.New("prevent further processing")
	}
}

// Verify a specific method is supported.
func (h *Handler) isSupported(method string) bool {
	for _, m := range h.methods {
		if method == m {
			return true
		}
	}
	return false
}
