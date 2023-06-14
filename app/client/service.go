package client

import (
	"context"

	"google.golang.org/grpc"
)

type ServiceGrpcClientConn struct {
	CC grpc.ClientConnInterface
}

func WrapServiceGrpcClientConn(cc grpc.ClientConnInterface) *ServiceGrpcClientConn {
	return &ServiceGrpcClientConn{
		CC: cc,
	}
}

func (gcc *ServiceGrpcClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	_, ok := ctx.Deadline()
	if !ok {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, Timeout)
		defer cancel()
	}
	return gcc.CC.Invoke(ctx, method, args, reply, opts...)
}

func (gcc *ServiceGrpcClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return gcc.CC.NewStream(ctx, desc, method, opts...)
}
