package grpcx

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpcdirector"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpcoptions"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpcproxy"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/tracing"
	netutil "git.bianfeng.com/stars/wegame/wan/wanx/pkg/netx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/syncx"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	_ "git.bianfeng.com/stars/wegame/wan/wanx/grpcx/balancer"
)

const grpcTimeoutDefault = 5

// Factory 构建GRPC客户端和服务端
type Factory struct {
	options      grpcoptions.Options
	traceFactory *tracing.Factory

	ConnsMap syncx.Map[string, *grpc.ClientConn]
}

// RuntimeServer 接收来自APP的调用和GRPC代理请求
func (f *Factory) RuntimeServer(ush grpc.StreamHandler, addOpts ...grpc.ServerOption) (*grpc.Server, error) {
	tp, err := f.traceFactory.NewTracerProvider(attribute.String("sgr.kind", "runtime-server"))
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{
		// 大文件支持
		grpc.MaxSendMsgSize(f.options.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(f.options.MaxRecvMsgSize),
		// 提高吞吐量
		grpc.InitialWindowSize(int32(f.options.InitialWindowSize)),
		grpc.InitialConnWindowSize(int32(f.options.InitialConnWindowSize)),
		grpc.MaxConcurrentStreams(uint32(f.options.MaxConcurrentStreams)),

		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				// grpc_zap.StreamServerInterceptor(zapLogger),
				otelgrpc.StreamServerInterceptor(
					otelgrpc.WithTracerProvider(tp),
					otelgrpc.WithInterceptorFilter(nil),
				),
				grpc_recovery.StreamServerInterceptor(),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				// grpc_zap.UnaryServerInterceptor(zapLogger),
				otelgrpc.UnaryServerInterceptor(
					otelgrpc.WithTracerProvider(tp),
					otelgrpc.WithInterceptorFilter(nil),
				),
				grpc_recovery.UnaryServerInterceptor(),
			),
		),
		grpc.UnknownServiceHandler(ush),
	}
	server := grpc.NewServer(append(opts, addOpts...)...)
	return server, nil
}

// RelayClient 转发 RuntimeServer 收到的请求，调用RelayServer
func (f *Factory) RelayClient(serviceName, schema string, defaultServiceConfig string) (*grpc.ClientConn, error) {
	tp, err := f.traceFactory.NewTracerProvider(attribute.String("sgr.kind", "relay-client"))
	if err != nil {
		return nil, err
	}
	if conn, ok := f.ConnsMap.Load(serviceName); ok {
		return conn, nil
	}
	if defaultServiceConfig == "" {
		defaultServiceConfig = `{"loadBalancingConfig": [{"round_robin":{}}]}`
	}

	// 构建target,这里可能会冲突
	var target strings.Builder
	target.WriteString("relay://")
	target.WriteString(serviceName)
	target.WriteByte('/')
	target.WriteString("")
	target.WriteString("?schema=")
	target.WriteString(url.QueryEscape(schema))
	ctx, cancel := context.WithTimeout(context.TODO(), grpcTimeoutDefault*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx,
		target.String(),
		// 设置默认负载均衡调度算法为轮询
		// see https://github.com/grpc/grpc/blob/master/doc/service_config.md
		grpc.WithDefaultServiceConfig(defaultServiceConfig),
		// 优化GRPC吞吐量
		grpc.WithInitialWindowSize(int32(f.options.InitialWindowSize)),
		grpc.WithInitialConnWindowSize(int32(f.options.InitialConnWindowSize)),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(false),
			grpc.MaxCallSendMsgSize(f.options.MaxSendMsgSize),
			grpc.MaxCallRecvMsgSize(f.options.MaxRecvMsgSize),
		),
		// 暂时不支持加密连接
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 调用链追踪
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			otelgrpc.UnaryClientInterceptor(
				otelgrpc.WithTracerProvider(tp),
				otelgrpc.WithInterceptorFilter(nil),
			),
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			otelgrpc.StreamClientInterceptor(
				otelgrpc.WithTracerProvider(tp),
				otelgrpc.WithInterceptorFilter(nil),
			),
		)),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "拨号错误,target=%s,err=%s", target.String(), err)
	}
	// 保存客户端连接
	f.ConnsMap.Store(serviceName, conn)
	return conn, nil
}

// RelayServer 收到 RelayClient 的调用，转发给 AppClient
func (f *Factory) RelayServer(appName, appAddr string) (*grpc.Server, error) {
	upstream, err := f.AppClient(appName, appAddr)
	if err != nil {
		return nil, err
	}
	ps, err := grpcdirector.NewReverseProxyDirector(appName, upstream)
	if err != nil {
		return nil, err
	}

	tpRelayServer, err := f.traceFactory.NewTracerProvider(attribute.String("sgr.kind", "relay-server"))
	if err != nil {
		return nil, err
	}
	// zapLogger, _ := zap.NewDevelopment()
	// zapLogger = zapLogger.Named("sgr-relay").WithOptions(zap.AddCallerSkip(2))
	server := grpc.NewServer(
		// 大文件支持
		grpc.MaxSendMsgSize(f.options.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(f.options.MaxRecvMsgSize),
		// 提高吞吐量
		grpc.InitialWindowSize(int32(f.options.InitialWindowSize)),
		grpc.InitialConnWindowSize(int32(f.options.InitialConnWindowSize)),
		grpc.MaxConcurrentStreams(uint32(f.options.MaxConcurrentStreams)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(tpRelayServer)),
				grpc_recovery.StreamServerInterceptor(),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tpRelayServer)),
				grpc_recovery.UnaryServerInterceptor(),
			),
		),
		grpc.UnknownServiceHandler(grpcproxy.TransparentHandler(ps.Director)),
	)
	return server, nil
}

// AppClient 接收RelayServer收到的请求，转发给APP
func (f *Factory) AppClient(appName, target string) (*grpc.ClientConn, error) {
	tp, err := f.traceFactory.NewTracerProvider(attribute.String("sgr.kind", "app-client"))
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(target,
		grpc.WithInitialWindowSize(1*1024*1024),
		grpc.WithInitialConnWindowSize(1*1024*1024),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(4*1024*1024), grpc.MaxCallRecvMsgSize(4*1024*1024)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
		grpc.WithUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(
				otelgrpc.WithTracerProvider(tp),
				otelgrpc.WithInterceptorFilter(nil),
			),
		),
		grpc.WithStreamInterceptor(
			otelgrpc.StreamClientInterceptor(
				otelgrpc.WithTracerProvider(tp),
				otelgrpc.WithInterceptorFilter(nil),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func dialer(ctx context.Context, addr string) (net.Conn, error) {
	if deadline, ok := ctx.Deadline(); ok {
		return netutil.DialTimeout(addr, time.Until(deadline))
	}
	return netutil.Dial(addr)
}
