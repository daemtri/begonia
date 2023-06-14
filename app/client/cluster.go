package client

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/balancer/specify"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClusterGrpcClientConn struct {
	specify string
	CC      grpc.ClientConnInterface
}

func WrapClusterGrpcClientConn(cc grpc.ClientConnInterface, instanceID string) *ClusterGrpcClientConn {
	return &ClusterGrpcClientConn{
		specify: "id=" + instanceID,
		CC:      cc,
	}
}

func (gcc *ClusterGrpcClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	_, ok := ctx.Deadline()
	if !ok {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, Timeout)
		defer cancel()

	}
	ctx2 := metadata.AppendToOutgoingContext(ctx, specify.PolicyMetadataKey, gcc.specify)
	return gcc.CC.Invoke(ctx2, method, args, reply, opts...)
}

func (gcc *ClusterGrpcClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx2 := metadata.AppendToOutgoingContext(ctx, specify.PolicyMetadataKey, gcc.specify)
	return gcc.CC.NewStream(ctx2, desc, method, opts...)
}
