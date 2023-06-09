package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"git.bianfeng.com/stars/wegame/wan/wanx/api/transmit"
	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap/header"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"google.golang.org/grpc"
)

type LogicServerRunOption struct {
	Addr string `flag:"addr" default:"0.0.0.0:8090" usage:"Grpc服务监听地址"`
}

// LogicServer 业务逻辑服务
type LogicServer struct {
	GrpcServer

	opt *LogicServerRunOption
	sb  *grpcx.ServerBuilder
	reg *ServiceRegistrar
	ci  *ContextInjector
	bs  *BusinessService
}

func NewLogicServer(opt *LogicServerRunOption, sb *grpcx.ServerBuilder, reg *ServiceRegistrar, bs *BusinessService, ci *ContextInjector) (*LogicServer, error) {
	ls := &LogicServer{
		opt: opt,
		sb:  sb,
		ci:  ci,
		bs:  bs,
		reg: reg,
	}
	return ls, ls.init()
}

func (ls *LogicServer) init() error {
	server, err := ls.sb.NewGrpcServer(nil, []grpc.UnaryServerInterceptor{
		header.MetadataInterceptor,
		ls.ci.Intercept,
	})
	if err != nil {
		return err
	}
	ls.reg.RegisterTo(server)
	ls.GrpcServer.Init(ls.opt.Addr, server)
	transmit.RegisterBusinessServiceServer(ls.server, ls.bs)
	return nil
}

func (ls *LogicServer) Enabled() bool {
	return len(ls.GrpcServer.server.GetServiceInfo()) > 1 || len(ls.reg.services) > 0
}

type ContextInjector struct {
	services map[string]func(ctx context.Context) context.Context
}

func NewContextInjector() (*ContextInjector, error) {
	return &ContextInjector{
		services: make(map[string]func(ctx context.Context) context.Context),
	}, nil
}

func (ci ContextInjector) Bind(serviceName string, injectFunc func(ctx context.Context) context.Context) {
	if _, ok := ci.services[serviceName]; ok {
		panic(fmt.Errorf("%s already bind a inject func", serviceName))
	}
	ci.services[serviceName] = injectFunc
}

func (ci ContextInjector) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	paths := strings.SplitN(info.FullMethod, "/", 3)
	if inject, ok := ci.services[paths[1]]; ok {
		return handler(inject(ctx), req)
	}
	return handler(ctx, req)
}
