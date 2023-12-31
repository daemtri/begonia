// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.1
// source: api/transmit/transmit.proto

package transmit

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BusinessServiceClient is the client API for BusinessService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BusinessServiceClient interface {
	Dispatch(ctx context.Context, in *DispatchRequest, opts ...grpc.CallOption) (*DispatchReply, error)
}

type businessServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBusinessServiceClient(cc grpc.ClientConnInterface) BusinessServiceClient {
	return &businessServiceClient{cc}
}

func (c *businessServiceClient) Dispatch(ctx context.Context, in *DispatchRequest, opts ...grpc.CallOption) (*DispatchReply, error) {
	out := new(DispatchReply)
	err := c.cc.Invoke(ctx, "/transmit.BusinessService/Dispatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BusinessServiceServer is the server API for BusinessService service.
// All implementations must embed UnimplementedBusinessServiceServer
// for forward compatibility
type BusinessServiceServer interface {
	Dispatch(context.Context, *DispatchRequest) (*DispatchReply, error)
	mustEmbedUnimplementedBusinessServiceServer()
}

// UnimplementedBusinessServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBusinessServiceServer struct {
}

func (UnimplementedBusinessServiceServer) Dispatch(context.Context, *DispatchRequest) (*DispatchReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Dispatch not implemented")
}
func (UnimplementedBusinessServiceServer) mustEmbedUnimplementedBusinessServiceServer() {}

// UnsafeBusinessServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BusinessServiceServer will
// result in compilation errors.
type UnsafeBusinessServiceServer interface {
	mustEmbedUnimplementedBusinessServiceServer()
}

func RegisterBusinessServiceServer(s grpc.ServiceRegistrar, srv BusinessServiceServer) {
	s.RegisterService(&BusinessService_ServiceDesc, srv)
}

func _BusinessService_Dispatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DispatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BusinessServiceServer).Dispatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/transmit.BusinessService/Dispatch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BusinessServiceServer).Dispatch(ctx, req.(*DispatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BusinessService_ServiceDesc is the grpc.ServiceDesc for BusinessService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BusinessService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "transmit.BusinessService",
	HandlerType: (*BusinessServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Dispatch",
			Handler:    _BusinessService_Dispatch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/transmit/transmit.proto",
}

// GatewayControlServiceClient is the client API for GatewayControlService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GatewayControlServiceClient interface {
	Notify(ctx context.Context, in *NotifyRequest, opts ...grpc.CallOption) (*Empty, error)
	BroadCast(ctx context.Context, in *BroadCastRequest, opts ...grpc.CallOption) (*Empty, error)
	Kick(ctx context.Context, in *KickRequest, opts ...grpc.CallOption) (*Empty, error)
}

type gatewayControlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGatewayControlServiceClient(cc grpc.ClientConnInterface) GatewayControlServiceClient {
	return &gatewayControlServiceClient{cc}
}

func (c *gatewayControlServiceClient) Notify(ctx context.Context, in *NotifyRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/transmit.GatewayControlService/Notify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayControlServiceClient) BroadCast(ctx context.Context, in *BroadCastRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/transmit.GatewayControlService/BroadCast", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayControlServiceClient) Kick(ctx context.Context, in *KickRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/transmit.GatewayControlService/Kick", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GatewayControlServiceServer is the server API for GatewayControlService service.
// All implementations must embed UnimplementedGatewayControlServiceServer
// for forward compatibility
type GatewayControlServiceServer interface {
	Notify(context.Context, *NotifyRequest) (*Empty, error)
	BroadCast(context.Context, *BroadCastRequest) (*Empty, error)
	Kick(context.Context, *KickRequest) (*Empty, error)
	mustEmbedUnimplementedGatewayControlServiceServer()
}

// UnimplementedGatewayControlServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGatewayControlServiceServer struct {
}

func (UnimplementedGatewayControlServiceServer) Notify(context.Context, *NotifyRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Notify not implemented")
}
func (UnimplementedGatewayControlServiceServer) BroadCast(context.Context, *BroadCastRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadCast not implemented")
}
func (UnimplementedGatewayControlServiceServer) Kick(context.Context, *KickRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Kick not implemented")
}
func (UnimplementedGatewayControlServiceServer) mustEmbedUnimplementedGatewayControlServiceServer() {}

// UnsafeGatewayControlServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GatewayControlServiceServer will
// result in compilation errors.
type UnsafeGatewayControlServiceServer interface {
	mustEmbedUnimplementedGatewayControlServiceServer()
}

func RegisterGatewayControlServiceServer(s grpc.ServiceRegistrar, srv GatewayControlServiceServer) {
	s.RegisterService(&GatewayControlService_ServiceDesc, srv)
}

func _GatewayControlService_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayControlServiceServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/transmit.GatewayControlService/Notify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayControlServiceServer).Notify(ctx, req.(*NotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GatewayControlService_BroadCast_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadCastRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayControlServiceServer).BroadCast(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/transmit.GatewayControlService/BroadCast",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayControlServiceServer).BroadCast(ctx, req.(*BroadCastRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GatewayControlService_Kick_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KickRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayControlServiceServer).Kick(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/transmit.GatewayControlService/Kick",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayControlServiceServer).Kick(ctx, req.(*KickRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GatewayControlService_ServiceDesc is the grpc.ServiceDesc for GatewayControlService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GatewayControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "transmit.GatewayControlService",
	HandlerType: (*GatewayControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Notify",
			Handler:    _GatewayControlService_Notify_Handler,
		},
		{
			MethodName: "BroadCast",
			Handler:    _GatewayControlService_BroadCast_Handler,
		},
		{
			MethodName: "Kick",
			Handler:    _GatewayControlService_Kick_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/transmit/transmit.proto",
}
