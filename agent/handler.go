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
	"github.com/algorandfoundation/did-algo/info"
	protov1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/did"
	xlog "go.bryk.io/pkg/log"
	"go.bryk.io/pkg/net/rpc"
	"go.bryk.io/pkg/otel"
	"google.golang.org/grpc"
)

// Handler provides the required functionality for the DID method
type Handler struct {
	oop         *otel.Operator
	methods     []string
	store       Storage
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

	// Observability operator.
	OOP *otel.Operator

	// Algorand node client.
	AlgoNode *algod.Client

	// Algorand indexer client.
	AlgoIndexer *indexer.Client
}

// NewHandler starts a new DID method handler instance
func NewHandler(options HandlerOptions) (*Handler, error) {
	return &Handler{
		oop:         options.OOP,
		store:       options.Store,
		methods:     options.Methods,
		difficulty:  options.Difficulty,
		algoNode:    options.AlgoNode,
		algoIndexer: options.AlgoIndexer,
	}, nil
}

// Close the instance and safely terminate any internal processing
func (h *Handler) Close() error {
	h.oop.Info("closing agent handler")
	return h.store.Close()
}

// Retrieve an existing DID instance based on its subject string
func (h *Handler) Retrieve(req *protov1.QueryRequest) (*did.Identifier, *did.ProofLD, error) {
	logFields := xlog.Fields{
		"method":  req.Method,
		"subject": req.Subject,
	}
	h.oop.WithFields(logFields).Debug("retrieve request")

	// Verify method is supported
	if !h.isSupported(req.Method) {
		h.oop.WithFields(logFields).Warning("non supported method")
		return nil, nil, errors.New("non supported method")
	}

	// Retrieve document from storage
	id, proof, err := h.store.Get(req)
	if err != nil {
		h.oop.WithFields(logFields).Warning(err.Error())
		return nil, nil, err
	}
	return id, proof, nil
}

// Process an incoming request ticket
func (h *Handler) Process(req *protov1.ProcessRequest) (string, error) {
	// Empty request
	if req == nil {
		return "", errors.New("empty request")
	}

	// Validate ticket
	if err := req.Ticket.Verify(h.difficulty); err != nil {
		h.oop.WithFields(xlog.Fields{"error": err.Error()}).Error("invalid ticket")
		return "", err
	}

	// Load DID document and proof
	id, err := req.Ticket.GetDID()
	if err != nil {
		h.oop.WithFields(xlog.Fields{"error": err.Error()}).Error("invalid DID contents")
		return "", err
	}
	proof, err := req.Ticket.GetProofLD()
	if err != nil {
		h.oop.WithFields(xlog.Fields{"error": err.Error()}).Error("invalid DID proof")
		return "", err
	}

	// Verify method is supported
	if !h.isSupported(id.Method()) {
		h.oop.WithFields(xlog.Fields{"method": id.Method()}).Warning("non supported method")
		return "", errors.New("non supported method")
	}

	// Update operations require another validation step using the original record
	isUpdate := h.store.Exists(id)
	if isUpdate {
		if err := req.Ticket.Verify(h.difficulty); err != nil {
			h.oop.WithFields(xlog.Fields{"error": err.Error()}).Error("invalid ticket")
			return "", err
		}
	}

	h.oop.WithFields(xlog.Fields{
		"subject": id.Subject(),
		"update":  isUpdate,
		"task":    req.Task,
	}).Debug("write operation")
	cid := ""
	cid, err = h.store.Save(id, proof)

	return cid, err
}

// AccountInformation returns details about the crypto account specified.
func (h *Handler) AccountInformation(
	ctx context.Context,
	req *protov1.AccountInformationRequest) (*protov1.AccountInformationResponse, error) {
	ai, err := h.algoNode.AccountInformation(req.Address).Do(ctx)
	if err != nil {
		h.oop.WithFields(xlog.Fields{
			"error":   err.Error(),
			"address": req.Address,
		}).Error("failed to get account information")
		return nil, err
	}
	_, ptList, err := h.algoNode.PendingTransactionsByAddress(req.Address).Do(ctx)
	if err != nil {
		h.oop.WithFields(xlog.Fields{
			"error":   err.Error(),
			"address": req.Address,
		}).Error("failed to get pending transactions")
		return nil, err
	}
	resp := &protov1.AccountInformationResponse{
		Status:              ai.Status,
		Balance:             ai.AmountWithoutPendingRewards,
		TotalRewards:        ai.Rewards,
		PendingRewards:      ai.PendingRewards,
		PendingTransactions: []*protov1.AlgoTransaction{},
	}
	for _, pt := range ptList {
		resp.PendingTransactions = append(resp.PendingTransactions, &protov1.AlgoTransaction{
			Amount:   uint64(pt.Txn.Amount),
			Receiver: pt.Txn.Receiver.String(),
			Note:     pt.Txn.Note,
		})
	}
	return resp, nil
}

// AccountActivity opens a channel to monitor near real-time account activity.
// The channel must be closed using the provided context when no longer needed.
func (h *Handler) AccountActivity(
	ctx context.Context,
	req *protov1.AccountActivityRequest) (<-chan *protov1.AccountActivityResponse, error) {
	check := time.NewTicker(time.Duration(5) * time.Second)
	sink := make(chan *protov1.AccountActivityResponse)
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
					h.oop.WithFields(xlog.Fields{
						"error":   err.Error(),
						"address": req.Address,
					}).Error("failed to get account activity")
				} else {
					sink <- &protov1.AccountActivityResponse{
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
func (h *Handler) TxParameters(ctx context.Context) (*protov1.TxParametersResponse, error) {
	params, err := h.algoNode.SuggestedParams().Do(ctx)
	if err != nil {
		h.oop.WithFields(xlog.Fields{
			"error": err.Error(),
		}).Error("failed to get transaction parameters")
		return nil, err
	}
	data, err := json.Marshal(params)
	if err != nil {
		h.oop.WithFields(xlog.Fields{
			"error": err.Error(),
		}).Error("failed to encode transaction parameters")
		return nil, err
	}
	return &protov1.TxParametersResponse{Params: data}, nil
}

// TxSubmit will send a signed raw transaction to the network for processing.
func (h *Handler) TxSubmit(ctx context.Context, req *protov1.TxSubmitRequest) (*protov1.TxSubmitResponse, error) {
	tid, err := h.algoNode.SendRawTransaction(req.Stx).Do(ctx)
	if err != nil {
		h.oop.WithFields(xlog.Fields{
			"error": err.Error(),
			"tx":    base64.StdEncoding.EncodeToString(req.Stx),
		}).Error("failed to submit raw transaction")
		return nil, err
	}
	return &protov1.TxSubmitResponse{Id: tid}, nil
}

// ServiceDefinition allows the handler instance to be exposed using an RPC server
func (h *Handler) ServiceDefinition() *rpc.Service {
	return &rpc.Service{
		GatewaySetup: protov1.RegisterAgentAPIHandlerFromEndpoint,
		ServerSetup: func(s *grpc.Server) {
			protov1.RegisterAgentAPIServer(s, &rpcHandler{handler: h})
		},
	}
}

// QueryResponseFilter provides custom encoding of HTTP query results.
func (h *Handler) QueryResponseFilter() rpc.HTTPGatewayFilter {
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
		rr := &protov1.QueryRequest{
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
		res.Header().Set("content-type", "application/json")
		res.Header().Set("x-content-type-options", "nosniff")
		res.Header().Set("x-algoid-build-code", info.BuildCode)
		res.Header().Set("x-algoid-build-timestamp", info.BuildTimestamp)
		res.Header().Set("x-algoid-version", info.CoreVersion)
		res.WriteHeader(status)
		_, _ = res.Write(response)
		return errors.New("prevent further processing")
	}
}

// Verify a specific method is supported
func (h *Handler) isSupported(method string) bool {
	for _, m := range h.methods {
		if method == m {
			return true
		}
	}
	return false
}
