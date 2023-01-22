package agent

import (
	"context"
	"encoding/json"

	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Wrapper to enable RPC access to an underlying method handler instance.
type rpcHandler struct {
	protoV1.UnimplementedAgentAPIServer
	handler *Handler
}

func (rh *rpcHandler) Ping(ctx context.Context, _ *emptypb.Empty) (*protoV1.PingResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Ping", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(nil)
	return &protoV1.PingResponse{Ok: true}, nil
}

func (rh *rpcHandler) Process(ctx context.Context, req *protoV1.ProcessRequest) (res *protoV1.ProcessResponse, err error) { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Process", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(err)

	// Process and return response
	res.Identifier, err = rh.handler.Process(req)
	res.Ok = err == nil
	if err != nil {
		err = status.Error(codes.InvalidArgument, err.Error())
	}
	return
}

func (rh *rpcHandler) Query(ctx context.Context, req *protoV1.QueryRequest) (res *protoV1.QueryResponse, err error) { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Query", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(err)

	// Process and return response
	id, proof, err := rh.handler.Retrieve(req)
	if err != nil {
		err = status.Error(codes.NotFound, err.Error())
		return
	}
	res.Proof, _ = json.Marshal(proof)
	res.Document, _ = json.Marshal(id.Document(true))
	res.DocumentMetadata, _ = json.Marshal(id.GetMetadata())
	return
}

func (rh *rpcHandler) TxParameters(ctx context.Context, _ *emptypb.Empty) (res *protoV1.TxParametersResponse, err error) { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(ctx, "TxParameters", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(err)
	return rh.handler.TxParameters(ctx)
}

func (rh *rpcHandler) TxSubmit(ctx context.Context, req *protoV1.TxSubmitRequest) (res *protoV1.TxSubmitResponse, err error) { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(ctx, "TxSubmit", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(err)
	return rh.handler.TxSubmit(ctx, req)
}

func (rh *rpcHandler) AccountInformation(ctx context.Context, req *protoV1.AccountInformationRequest) (res *protoV1.AccountInformationResponse, err error) { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(ctx, "AccountInformation", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End(err)
	return rh.handler.AccountInformation(ctx, req)
}

func (rh *rpcHandler) AccountActivity(req *protoV1.AccountActivityRequest, stream protoV1.AgentAPI_AccountActivityServer) error { // nolint: lll
	// Track operation
	sp := rh.handler.oop.Start(stream.Context(), "AccountActivity", otel.WithSpanKind(otel.SpanKindServer))

	// Open account monitor
	monitor, err := rh.handler.AccountActivity(stream.Context(), req)
	if err != nil {
		sp.End(err)
		return err
	}

	// Stream account activity
	for record := range monitor {
		if err = stream.Send(record); err != nil {
			sp.End(err)
			return err
		}
	}
	sp.End(nil)
	return nil
}
