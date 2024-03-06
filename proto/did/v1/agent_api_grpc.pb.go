// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             buf-v1.29.0
// source: did/v1/agent_api.proto

package didv1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	AgentAPI_Ping_FullMethodName               = "/did.v1.AgentAPI/Ping"
	AgentAPI_Process_FullMethodName            = "/did.v1.AgentAPI/Process"
	AgentAPI_Query_FullMethodName              = "/did.v1.AgentAPI/Query"
	AgentAPI_AccountInformation_FullMethodName = "/did.v1.AgentAPI/AccountInformation"
	AgentAPI_TxParameters_FullMethodName       = "/did.v1.AgentAPI/TxParameters"
	AgentAPI_TxSubmit_FullMethodName           = "/did.v1.AgentAPI/TxSubmit"
	AgentAPI_AccountActivity_FullMethodName    = "/did.v1.AgentAPI/AccountActivity"
)

// AgentAPIClient is the client API for AgentAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AgentAPIClient interface {
	// Reachability test.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PingResponse, error)
	// Process an incoming request ticket.
	Process(ctx context.Context, in *ProcessRequest, opts ...grpc.CallOption) (*ProcessResponse, error)
	// Return the current state of a DID subject.
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error)
	// Request information about an Algorand account.
	AccountInformation(ctx context.Context, in *AccountInformationRequest, opts ...grpc.CallOption) (*AccountInformationResponse, error)
	// Return the current transaction parameters for the network.
	TxParameters(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TxParametersResponse, error)
	// Submit a raw signed transaction to the network for processing.
	TxSubmit(ctx context.Context, in *TxSubmitRequest, opts ...grpc.CallOption) (*TxSubmitResponse, error)
	// Provide near real-time notifications for account activity.
	AccountActivity(ctx context.Context, in *AccountActivityRequest, opts ...grpc.CallOption) (AgentAPI_AccountActivityClient, error)
}

type agentAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentAPIClient(cc grpc.ClientConnInterface) AgentAPIClient {
	return &agentAPIClient{cc}
}

func (c *agentAPIClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, AgentAPI_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) Process(ctx context.Context, in *ProcessRequest, opts ...grpc.CallOption) (*ProcessResponse, error) {
	out := new(ProcessResponse)
	err := c.cc.Invoke(ctx, AgentAPI_Process_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := c.cc.Invoke(ctx, AgentAPI_Query_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) AccountInformation(ctx context.Context, in *AccountInformationRequest, opts ...grpc.CallOption) (*AccountInformationResponse, error) {
	out := new(AccountInformationResponse)
	err := c.cc.Invoke(ctx, AgentAPI_AccountInformation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) TxParameters(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TxParametersResponse, error) {
	out := new(TxParametersResponse)
	err := c.cc.Invoke(ctx, AgentAPI_TxParameters_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) TxSubmit(ctx context.Context, in *TxSubmitRequest, opts ...grpc.CallOption) (*TxSubmitResponse, error) {
	out := new(TxSubmitResponse)
	err := c.cc.Invoke(ctx, AgentAPI_TxSubmit_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentAPIClient) AccountActivity(ctx context.Context, in *AccountActivityRequest, opts ...grpc.CallOption) (AgentAPI_AccountActivityClient, error) {
	stream, err := c.cc.NewStream(ctx, &AgentAPI_ServiceDesc.Streams[0], AgentAPI_AccountActivity_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &agentAPIAccountActivityClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AgentAPI_AccountActivityClient interface {
	Recv() (*AccountActivityResponse, error)
	grpc.ClientStream
}

type agentAPIAccountActivityClient struct {
	grpc.ClientStream
}

func (x *agentAPIAccountActivityClient) Recv() (*AccountActivityResponse, error) {
	m := new(AccountActivityResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// AgentAPIServer is the server API for AgentAPI service.
// All implementations must embed UnimplementedAgentAPIServer
// for forward compatibility
type AgentAPIServer interface {
	// Reachability test.
	Ping(context.Context, *emptypb.Empty) (*PingResponse, error)
	// Process an incoming request ticket.
	Process(context.Context, *ProcessRequest) (*ProcessResponse, error)
	// Return the current state of a DID subject.
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
	// Request information about an Algorand account.
	AccountInformation(context.Context, *AccountInformationRequest) (*AccountInformationResponse, error)
	// Return the current transaction parameters for the network.
	TxParameters(context.Context, *emptypb.Empty) (*TxParametersResponse, error)
	// Submit a raw signed transaction to the network for processing.
	TxSubmit(context.Context, *TxSubmitRequest) (*TxSubmitResponse, error)
	// Provide near real-time notifications for account activity.
	AccountActivity(*AccountActivityRequest, AgentAPI_AccountActivityServer) error
	mustEmbedUnimplementedAgentAPIServer()
}

// UnimplementedAgentAPIServer must be embedded to have forward compatible implementations.
type UnimplementedAgentAPIServer struct {
}

func (UnimplementedAgentAPIServer) Ping(context.Context, *emptypb.Empty) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedAgentAPIServer) Process(context.Context, *ProcessRequest) (*ProcessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Process not implemented")
}
func (UnimplementedAgentAPIServer) Query(context.Context, *QueryRequest) (*QueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}
func (UnimplementedAgentAPIServer) AccountInformation(context.Context, *AccountInformationRequest) (*AccountInformationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountInformation not implemented")
}
func (UnimplementedAgentAPIServer) TxParameters(context.Context, *emptypb.Empty) (*TxParametersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxParameters not implemented")
}
func (UnimplementedAgentAPIServer) TxSubmit(context.Context, *TxSubmitRequest) (*TxSubmitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxSubmit not implemented")
}
func (UnimplementedAgentAPIServer) AccountActivity(*AccountActivityRequest, AgentAPI_AccountActivityServer) error {
	return status.Errorf(codes.Unimplemented, "method AccountActivity not implemented")
}
func (UnimplementedAgentAPIServer) mustEmbedUnimplementedAgentAPIServer() {}

// UnsafeAgentAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AgentAPIServer will
// result in compilation errors.
type UnsafeAgentAPIServer interface {
	mustEmbedUnimplementedAgentAPIServer()
}

func RegisterAgentAPIServer(s grpc.ServiceRegistrar, srv AgentAPIServer) {
	s.RegisterService(&AgentAPI_ServiceDesc, srv)
}

func _AgentAPI_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_Process_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProcessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).Process(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_Process_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).Process(ctx, req.(*ProcessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_Query_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_AccountInformation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccountInformationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).AccountInformation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_AccountInformation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).AccountInformation(ctx, req.(*AccountInformationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_TxParameters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).TxParameters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_TxParameters_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).TxParameters(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_TxSubmit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxSubmitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentAPIServer).TxSubmit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentAPI_TxSubmit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentAPIServer).TxSubmit(ctx, req.(*TxSubmitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentAPI_AccountActivity_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(AccountActivityRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AgentAPIServer).AccountActivity(m, &agentAPIAccountActivityServer{stream})
}

type AgentAPI_AccountActivityServer interface {
	Send(*AccountActivityResponse) error
	grpc.ServerStream
}

type agentAPIAccountActivityServer struct {
	grpc.ServerStream
}

func (x *agentAPIAccountActivityServer) Send(m *AccountActivityResponse) error {
	return x.ServerStream.SendMsg(m)
}

// AgentAPI_ServiceDesc is the grpc.ServiceDesc for AgentAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AgentAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "did.v1.AgentAPI",
	HandlerType: (*AgentAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _AgentAPI_Ping_Handler,
		},
		{
			MethodName: "Process",
			Handler:    _AgentAPI_Process_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _AgentAPI_Query_Handler,
		},
		{
			MethodName: "AccountInformation",
			Handler:    _AgentAPI_AccountInformation_Handler,
		},
		{
			MethodName: "TxParameters",
			Handler:    _AgentAPI_TxParameters_Handler,
		},
		{
			MethodName: "TxSubmit",
			Handler:    _AgentAPI_TxSubmit_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "AccountActivity",
			Handler:       _AgentAPI_AccountActivity_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "did/v1/agent_api.proto",
}
