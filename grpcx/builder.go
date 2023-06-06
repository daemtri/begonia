package grpcx

import (
	"context"
	"flag"
	"net/url"
	"strings"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpclogx"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpcoptions"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/grpcresolver"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/tracing"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

func init() {
	grpclog.SetLoggerV2(grpclogx.NewGrpcLog(logx.GetLogger("grpc").Handler()))
}

type FactoryBuilder struct {
	options grpcoptions.Options
}

func NewFactoryBuilder() *FactoryBuilder {
	return &FactoryBuilder{options: grpcoptions.NewOptions()}
}

func (f *FactoryBuilder) AddFlags(fs *flag.FlagSet) {
	f.options.AddFlags(fs)
}

func (f *FactoryBuilder) ValidateFlags() error {
	return f.options.ValidateFlags()
}

func (f *FactoryBuilder) Build(ctx context.Context) (*Factory, error) {
	discovery := box.Invoke[component.Discovery](ctx)
	grpcresolver.RegisterInCluster("relay", discovery)
	return &Factory{
		options:      f.options,
		traceFactory: box.Invoke[*tracing.Factory](ctx),
	}, nil
}

type ServerOptions struct {
	// MaxSendSize GRPC服务参数
	MaxSendMsgSize        int  `flag:"max_send_msg_size" default:"4194304" usage:""`        // 4 * 1024 * 1024
	MaxRecvMsgSize        int  `flag:"max_recv_msg_size" default:"4194304" usage:""`        //  4 * 1024 * 1024
	InitialWindowSize     int  `flag:"initial_window_size" default:"1048576" usage:""`      //  1 * 1024 * 1024
	InitialConnWindowSize int  `flag:"initial_conn_window_size" default:"1048576" usage:""` //  1 * 1024 * 1024
	MaxConcurrentStreams  uint `flag:"max_concurrent_streams" default:"10000" usage:""`
}

type ServerBuilder struct {
	opts *ServerOptions
	tb   *tracing.Factory
}

func NewServerBuilder(opts *ServerOptions, tb *tracing.Factory) (*ServerBuilder, error) {
	return &ServerBuilder{opts: opts, tb: tb}, nil
}

func (sb *ServerBuilder) NewGrpcServer(streamInterceptors []grpc.StreamServerInterceptor, unaryInterceptors []grpc.UnaryServerInterceptor) (*grpc.Server, error) {
	tp, err := sb.tb.NewTracerProvider(attribute.String("sgr.kind", "runtime-server"))
	if err != nil {
		return nil, err
	}
	streamInterceptors = append(
		streamInterceptors,
		grpc_ctxtags.StreamServerInterceptor(),
		// grpc_zap.StreamServerInterceptor(zapLogger),
		otelgrpc.StreamServerInterceptor(
			otelgrpc.WithTracerProvider(tp),
			otelgrpc.WithInterceptorFilter(nil),
		),
		grpc_recovery.StreamServerInterceptor(),
	)

	unaryInterceptors = append(
		unaryInterceptors,
		grpc_ctxtags.UnaryServerInterceptor(),
		// grpc_zap.UnaryServerInterceptor(zapLogger),
		otelgrpc.UnaryServerInterceptor(
			otelgrpc.WithTracerProvider(tp),
			otelgrpc.WithInterceptorFilter(nil),
		),
		grpc_recovery.UnaryServerInterceptor(),
	)

	opts := []grpc.ServerOption{
		// 大文件支持
		grpc.MaxSendMsgSize(sb.opts.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(sb.opts.MaxRecvMsgSize),
		// 提高吞吐量
		grpc.InitialWindowSize(int32(sb.opts.InitialWindowSize)),
		grpc.InitialConnWindowSize(int32(sb.opts.InitialConnWindowSize)),
		grpc.MaxConcurrentStreams(uint32(sb.opts.MaxConcurrentStreams)),

		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
	}
	server := grpc.NewServer(opts...)
	return server, nil
}

type ClientOptions struct {
	MaxSendMsgSize        int `flag:"max_send_msg_size" default:"4194304" usage:""`        // 4 * 1024 * 1024
	MaxRecvMsgSize        int `flag:"max_recv_msg_size" default:"4194304" usage:""`        //  4 * 1024 * 1024
	InitialWindowSize     int `flag:"initial_window_size" default:"1048576" usage:""`      //  1 * 1024 * 1024
	InitialConnWindowSize int `flag:"initial_conn_window_size" default:"1048576" usage:""` //  1 * 1024 * 1024
}

type ClientBuilder struct {
	opts         *ClientOptions
	traceFactory *tracing.Factory
}

func NewClientBuilder(opts *ClientOptions, tb *tracing.Factory, discovery component.Discovery) (*ClientBuilder, error) {
	grpcresolver.RegisterInCluster("relay", discovery)
	return &ClientBuilder{opts: opts, traceFactory: tb}, nil
}

func (cb *ClientBuilder) NewGrpcClientConn(serviceName string, schema string, defaultServiceConfig string) (grpc.ClientConnInterface, error) {
	tp, err := cb.traceFactory.NewTracerProvider(attribute.String("sgr.kind", "relay-client"))
	if err != nil {
		return nil, err
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
		grpc.WithInitialWindowSize(int32(cb.opts.InitialWindowSize)),
		grpc.WithInitialConnWindowSize(int32(cb.opts.InitialConnWindowSize)),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(false),
			grpc.MaxCallSendMsgSize(cb.opts.MaxSendMsgSize),
			grpc.MaxCallRecvMsgSize(cb.opts.MaxRecvMsgSize),
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

	return conn, nil
}
