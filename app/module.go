package app

import (
	"context"
	"fmt"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
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
			if err := mr.module.Init(withObjectContainer(ctx, mr)); err != nil {
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
				modules[i].module.Destroy(withObjectContainer(wCtx, mr))
			}
		}()
		return nil
	}
}

type Module interface {
	// Init 模块初始化
	Init(ctx context.Context) error
	// Integrate 注册业务服务处理器
	Integrate(ig Integrator)
	// Destroy
	Destroy(ctx context.Context) error
}

func RegisterModule[T Module](name string, m T) {
	if _, ok := modules[name]; ok {
		panic(fmt.Sprintf("module %s already registered", name))
	}
	modules[name] = m
}
