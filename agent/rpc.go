package agent

import (
	"context"
	"encoding/json"

	protoV1 "github.com/algorandfoundation/did-algo/proto/did/v1"
	"go.bryk.io/pkg/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Wrapper to enable RPC access to an underlying method handler instance.
type rpcHandler struct {
	protoV1.UnimplementedAgentAPIServer
	handler *Handler
}

func getHeaders() metadata.MD {
	return metadata.New(map[string]string{
		"x-content-type-options": "nosniff",
	})
}

func (rh *rpcHandler) Ping(ctx context.Context, _ *emptypb.Empty) (*protoV1.PingResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Ping", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()

	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &protoV1.PingResponse{Ok: true}, nil
}

func (rh *rpcHandler) Process(ctx context.Context, req *protoV1.ProcessRequest) (*protoV1.ProcessResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Process", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()

	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	cid, err := rh.handler.Process(req)
	if err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return &protoV1.ProcessResponse{Ok: false}, status.Error(codes.InvalidArgument, err.Error())
	}
	return &protoV1.ProcessResponse{
		Ok:         true,
		Identifier: cid,
	}, nil
}

func (rh *rpcHandler) Query(ctx context.Context, req *protoV1.QueryRequest) (*protoV1.QueryResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "Query", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()

	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	id, proof, err := rh.handler.Retrieve(req)
	if err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, err.Error())
	}
	doc, _ := json.Marshal(id.Document(true))
	pp, _ := json.Marshal(proof)
	documentMetadata, _ := json.Marshal(id.GetMetadata())
	return &protoV1.QueryResponse{
		Document:         doc,
		Proof:            pp,
		DocumentMetadata: documentMetadata,
	}, nil
}

func (rh *rpcHandler) TxParameters(ctx context.Context, _ *emptypb.Empty) (*protoV1.TxParametersResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "TxParameters", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()
	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return rh.handler.TxParameters(ctx)
}

func (rh *rpcHandler) TxSubmit(ctx context.Context, req *protoV1.TxSubmitRequest) (*protoV1.TxSubmitResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "TxSubmit", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()
	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return rh.handler.TxSubmit(ctx, req)
}

func (rh *rpcHandler) AccountInformation(
	ctx context.Context,
	req *protoV1.AccountInformationRequest) (*protoV1.AccountInformationResponse, error) {
	// Track operation
	sp := rh.handler.oop.Start(ctx, "AccountInformation", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()
	if err := grpc.SendHeader(ctx, getHeaders()); err != nil {
		sp.SetStatus(otelcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return rh.handler.AccountInformation(ctx, req)
}

func (rh *rpcHandler) AccountActivity(
	req *protoV1.AccountActivityRequest,
	stream protoV1.AgentAPI_AccountActivityServer) error {
	// Track operation
	sp := rh.handler.oop.Start(stream.Context(), "AccountActivity", otel.WithSpanKind(otel.SpanKindServer))
	defer sp.End()

	// Open account monitor
	monitor, err := rh.handler.AccountActivity(stream.Context(), req)
	if err != nil {
		return err
	}

	// Stream account activity
	for record := range monitor {
		if err = stream.Send(record); err != nil {
			return err
		}
	}
	return nil
}
