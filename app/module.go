package app

import (
	"context"
	"fmt"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	modules = make(map[string]Module)

	// globalIntegrator is the global Integrator for all modules
	globalIntegrator *Integrator
	currentModule    *moduleRuntime
)

func initModules() func(ctx context.Context) error {
	for name := range modules {
		box.Provide[*moduleRuntime](newModuleRuntime(name, modules[name]), box.WithFlags("module-"+name), box.WithName(name))
	}
	return func(ctx context.Context) error {
		if err := initGlobal(ctx); err != nil {
			return err
		}
		globalIntegrator = box.Invoke[*Integrator](ctx)

		modules := box.Invoke[[]*moduleRuntime](ctx)
		for i := range modules {
			mr := modules[i]
			if err := mr.module.Init(withModuleRuntime(ctx, mr)); err != nil {
				return err
			}
			globalIntegrator.integrate(mr)
		}
		go func() {
			<-ctx.Done()
			wCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for i := range modules {
				mr := modules[i]
				modules[i].module.Destroy(withModuleRuntime(wCtx, mr))
			}
		}()
		return nil
	}
}

type Module interface {
	// Init 模块初始化
	Init(ctx context.Context) error
	// Integrate 注册业务服务处理器
	Integrate(it Integrator)
	// Destroy
	Destroy(ctx context.Context) error
}

func RegisterModule[T Module](name string, m T) {
	if _, ok := modules[name]; ok {
		panic(fmt.Sprintf("module %s already registered", name))
	}
	modules[name] = m
}

type GrpcServiceRegistrar interface {
	contract.RouteRegistrar
	grpc.ServiceRegistrar
}

type grpcServiceRegistrarImpl struct {
	ci      *bootstrap.ContextInjector
	route   contract.RouteRegistrar
	service grpc.ServiceRegistrar
}

func newGrpcServiceRegistrarImpl(
	rr *bootstrap.RouteRegistrar,
	sr *bootstrap.ServiceRegistrar,
	ci *bootstrap.ContextInjector,
) (*grpcServiceRegistrarImpl, error) {
	lsri := &grpcServiceRegistrarImpl{
		route:   rr,
		service: sr,
		ci:      ci,
	}
	return lsri, nil
}

func (gr *grpcServiceRegistrarImpl) RegisterService(desc *grpc.ServiceDesc, impl any) {
	mr := currentModule
	gr.ci.Bind(desc.ServiceName, func(ctx context.Context) context.Context {
		return withModuleRuntime(ctx, mr)
	})
	gr.service.RegisterService(desc, impl)
}

func Route[K ~int32, T proto.Message](msgID K, handleFunc func(ctx context.Context, req T) error) contract.RouteCell {
	mr := currentModule
	return contract.RouteCell{
		MsgID: int32(msgID),
		HandleFunc: func(ctx context.Context, req []byte) error {
			var x T
			v := x.ProtoReflect().New().Interface()
			if req != nil {
				if err := proto.Unmarshal(req, v); err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("message id %d does not match the message type, error %s", msgID, err))
				}
			}
			return handleFunc(withModuleRuntime(ctx, mr), v.(T))
		},
	}
}

func (gr *grpcServiceRegistrarImpl) RegisterRoute(routes ...contract.RouteCell) {
	gr.route.RegisterRoute(routes...)
}

type httpServerMux struct {
	chi.Router
}

func newHttpServerMux() (*httpServerMux, error) {
	return &httpServerMux{
		Router: chi.NewRouter(),
	}, nil
}

func (mux *httpServerMux) Enabled() bool {
	logger.Info("http server enable check, routes", "routes", len(mux.Routes()))
	return len(mux.Routes()) > 0
}
