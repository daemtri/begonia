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

	// globalRegistry is the global registry for all modules
	globalRegistry    *registry
	currentModule     *moduleRuntime
	currentHttpRouter chi.Router
)

func initModules() func(ctx context.Context) error {
	for name := range modules {
		box.Provide[*moduleRuntime](newModuleRuntime(name, modules[name]), box.WithFlags("module-"+name), box.WithName(name))
	}
	return func(ctx context.Context) error {
		if err := initGlobal(ctx); err != nil {
			return err
		}
		globalRegistry = box.Invoke[*registry](ctx)

		modules := box.Invoke[[]*moduleRuntime](ctx)
		for i := range modules {
			mr := modules[i]
			if err := mr.module.Init(withModuleRuntime(ctx, mr)); err != nil {
				return err
			}
			currentModule = mr
			globalRegistry.http.Route("/"+mr.moduleName, func(r chi.Router) {
				currentHttpRouter = r
				mr.module.Integrate()
				currentHttpRouter = nil
			})
			currentModule = nil
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
	Integrate()
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
	route   contract.RouteRegistrar
	service grpc.ServiceRegistrar
}

func newGrpcServiceRegistrarImpl(rr *bootstrap.RouteRegistrar, sr *bootstrap.ServiceRegistrar) (*grpcServiceRegistrarImpl, error) {
	lsri := &grpcServiceRegistrarImpl{
		route:   rr,
		service: sr,
	}
	return lsri, nil
}

func (gr *grpcServiceRegistrarImpl) RegisterService(desc *grpc.ServiceDesc, impl any) {
	mr := currentModule
	globalRegistry.ci.Bind(desc.ServiceName, func(ctx context.Context) context.Context {
		return withModuleRuntime(ctx, mr)
	})
	gr.service.RegisterService(desc, impl)
}

func (gr *grpcServiceRegistrarImpl) RegisterRoute(msgID int32, handleFunc func(ctx context.Context, req []byte) error) {
	mr := currentModule
	gr.route.RegisterRoute(msgID, func(ctx context.Context, req []byte) error {
		return handleFunc(withModuleRuntime(ctx, mr), req)
	})
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
	return len(mux.Routes()) > 0
}

type registry struct {
	ci   *bootstrap.ContextInjector
	grpc GrpcServiceRegistrar
	http chi.Router

	contract.PubSubConsumerRegistrar
	contract.TaskProcessorRegistrar
}

func newRegistry(
	lsr GrpcServiceRegistrar,
	psr contract.PubSubConsumerRegistrar,
	tpr contract.TaskProcessorRegistrar,
	ci *bootstrap.ContextInjector,
	mux chi.Router,
) (*registry, error) {
	reg := &registry{
		grpc:                    lsr,
		PubSubConsumerRegistrar: psr,
		TaskProcessorRegistrar:  tpr,
		ci:                      ci,
		http:                    mux,
	}
	return reg, nil
}

func Http() chi.Router {
	return currentHttpRouter
}

func Grpc() GrpcServiceRegistrar {
	return globalRegistry.grpc
}

func RegisterRoute[K ~int32, T proto.Message](msgID K, handleFunc func(ctx context.Context, req T) error) {
	Grpc().RegisterRoute(int32(msgID), func(ctx context.Context, req []byte) error {
		var x T
		v := x.ProtoReflect().New().Interface()
		if req != nil {
			if err := proto.Unmarshal(req, v); err != nil {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("message id %d does not match the message type, error %s", msgID, err))
			}
		}
		return handleFunc(ctx, v.(T))
	})
}
