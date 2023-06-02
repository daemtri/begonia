package grpcdirector

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ReverseProxyDirector struct {
	serviceName string
	upstream    *grpc.ClientConn
}

func NewReverseProxyDirector(serviceName string, upstream *grpc.ClientConn) (*ReverseProxyDirector, error) {
	return &ReverseProxyDirector{
		serviceName: serviceName,
		upstream:    upstream,
	}, nil
}

func (rpd *ReverseProxyDirector) Director(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	outCtx := metadata.NewOutgoingContext(ctx, md.Copy())

	distAppName, _ := resolveAppName(fullMethodName, md)
	if distAppName != rpd.serviceName {
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}

	//log.Info().Str("dist-app-name", distAppName).Str("fullMethodName", fullMethodName).Msg("收到Internal Proxy请求")

	return outCtx, rpd.upstream, nil
}
