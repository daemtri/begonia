package grpcdirector

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	netutil "git.bianfeng.com/stars/wegame/wan/wanx/pkg/netx"

	"google.golang.org/grpc/metadata"
)

const (
	SgrAppNameMetaKey  = "sgr-app-name"
	SgrAppIDMetaKey    = "sgr-app-id"
	SgrBalancerMetaKey = "sgr-balancer"
)

var (
// log = sgrlog.Component("pkg.grpcx.grpcdirector")
)

// resolveAppName 从 fullMethodName 和 metadat.MD解析appName, appID可能为空
func resolveAppName(fullMethodName string, md metadata.MD) (appName, appID string) {
	instanceNames, ok := md[SgrAppIDMetaKey]
	if ok {
		appID = instanceNames[0]
	}
	// 首先从Metadata解析
	distAppNames, ok := md[SgrAppNameMetaKey]
	if ok {
		appName = distAppNames[0]
		return
	}
	// /mtx.sample.v1.Sample/TestUnary => [mtx.sample.v1.Sample, TestUnary]
	appServiceName, _, founded := strings.Cut(strings.TrimPrefix(fullMethodName, "/"), "/")
	if !founded {
		panic(fmt.Errorf("错误的fullMethodName: %s", fullMethodName))
	}
	// mtx.sample.v1.Sample => mtx.sample.v1
	pos := strings.LastIndex(appServiceName, ".")
	appName = appServiceName[:pos]
	return
}

func dialer(ctx context.Context, addr string) (net.Conn, error) {
	if deadline, ok := ctx.Deadline(); ok {
		return netutil.DialTimeout(addr, time.Until(deadline))
	}
	return netutil.Dial(addr)
}
