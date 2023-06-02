package grpcdirector

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ProxyDirector struct {
	clientFunc func(appName, balancer string) (*grpc.ClientConn, error)
}

func NewProxyDirector(clientFunc func(appName, balancer string) (*grpc.ClientConn, error)) *ProxyDirector {
	return &ProxyDirector{
		clientFunc: clientFunc,
	}
}

func (pd *ProxyDirector) Director(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	outCtx := metadata.NewOutgoingContext(ctx, md.Copy())

	var balancer string
	// 解析appName
	distAppName, appID := resolveAppName(fullMethodName, md)
	if appID != "" {
		balancer = "AppID." + appID
	} else {
		balances, ok := md[SgrBalancerMetaKey]
		if ok {
			balancer = balances[0]
		}
	}
	conn, err := pd.clientFunc(distAppName, balancer)

	return outCtx, conn, err
}
