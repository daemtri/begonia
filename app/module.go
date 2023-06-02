package app

import (
	"context"
	"fmt"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	modules = make(map[string]Module)
)

func initModules() func(ctx context.Context) error {
	for name := range modules {
		box.Provide[*moduleRuntime](newModuleRuntime(name, modules[name]), box.WithFlags("module-"+name), box.WithName(name))
	}
	return func(ctx context.Context) error {
		reg := box.Invoke[Registry](ctx)
		modules := box.Invoke[[]*moduleRuntime](ctx)
		for i := range modules {
			mr := modules[i]
			if err := mr.module.Init(withModuleRuntime(ctx, mr)); err != nil {
				return err
			}
			mr.module.Integrate(reg.clone(mr))
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
	Integrate(reg Registry)
	// Destroy
	Destroy(ctx context.Context) error
}

func RegisterModule[T Module](name string, m T) {
	if _, ok := modules[name]; ok {
		panic(fmt.Sprintf("module %s already registered", name))
	}
	modules[name] = m
}

type LogicServiceRegistrar interface {
	contract.RouteRegistrar
	grpc.ServiceRegistrar
}

type LogicServiceRegistrarImpl struct {
	contract.RouteRegistrar
	grpc.ServiceRegistrar
}

func newLogicServiceRegistrarImpl(rr *bootstrap.RouteRegistrar, sr *bootstrap.ServiceRegistrar) (*LogicServiceRegistrarImpl, error) {
	lsri := &LogicServiceRegistrarImpl{
		RouteRegistrar:   rr,
		ServiceRegistrar: sr,
	}
	return lsri, nil
}

type Registry struct {
	runtime *moduleRuntime
	ci      *bootstrap.ContextInjector
	logic   LogicServiceRegistrar

	contract.PubSubConsumerRegistrar
	contract.TaskProcessorRegistrar
}

func newRegistry(lsr LogicServiceRegistrar, psr contract.PubSubConsumerRegistrar, tpr contract.TaskProcessorRegistrar, ci *bootstrap.ContextInjector) (Registry, error) {
	reg := Registry{
		logic:                   lsr,
		PubSubConsumerRegistrar: psr,
		TaskProcessorRegistrar:  tpr,
		ci:                      ci,
	}
	return reg, nil
}

func (reg Registry) clone(mr *moduleRuntime) Registry {
	return Registry{
		runtime:                 mr,
		ci:                      reg.ci,
		logic:                   reg.logic,
		PubSubConsumerRegistrar: reg.PubSubConsumerRegistrar,
		TaskProcessorRegistrar:  reg.TaskProcessorRegistrar,
	}
}

func (reg Registry) RegisterService(desc *grpc.ServiceDesc, impl any) {
	reg.ci.Bind(desc.ServiceName, func(ctx context.Context) context.Context {
		return withModuleRuntime(ctx, reg.runtime)
	})
	reg.logic.RegisterService(desc, impl)
}

func RegisterRoute[K ~int32, T proto.Message](reg Registry, msgID K, handleFunc func(ctx context.Context, req T) error) {
	reg.logic.RegisterRoute(int32(msgID), func(ctx context.Context, req []byte) error {
		var x T
		v := x.ProtoReflect().New().Interface()
		if req != nil {
			if err := proto.Unmarshal(req, v); err != nil {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("message id %d does not match the message type, error %s", msgID, err))
			}
		}
		return handleFunc(withModuleRuntime(ctx, reg.runtime), v.(T))
	})
}

func RegisterService(reg Registry, desc *grpc.ServiceDesc, impl any) {
	reg.RegisterService(desc, impl)
}

func RegisterClient[T any](reg Registry, fn func(grpc.ClientConnInterface) T, name string) {

}
