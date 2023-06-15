package app

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/depency"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
)

type contextKey struct{ name string }

var (
	objectContainerCtxKey = &contextKey{name: "object_contianer"}
)

type objectContainer struct {
	*moduleRuntime
}

func withObjectContainer(ctx context.Context, mr *moduleRuntime) context.Context {
	return context.WithValue(ctx, objectContainerCtxKey, &objectContainer{
		moduleRuntime: mr,
	})
}

func objectContainerFromCtx(ctx context.Context) *objectContainer {
	v := ctx.Value(objectContainerCtxKey)
	if v == nil {
		panic("no object container in context")
	}
	return v.(*objectContainer)
}

type moduleOption struct {
	Dependecies []string `flag:"dependecies" usage:"依赖"`
	ConfigName  string   `flag:"config" usage:"配置名,默认为module_{module_name}"`
}

type moduleRuntime struct {
	moduleName string
	opts       *moduleOption
	module     Module
	config     helper.OnceCell[any]
}

func (mr *moduleRuntime) init() error {
	depency.SetModuleConfig(mr.moduleName, mr.opts.Dependecies)
	if mr.opts.ConfigName == "" {
		mr.opts.ConfigName = "module_" + mr.moduleName
	}
	return nil
}

func newModuleRuntime(name string, module Module) func(opts *moduleOption) (*moduleRuntime, error) {
	return func(opts *moduleOption) (*moduleRuntime, error) {
		mr := &moduleRuntime{
			moduleName: name,
			opts:       opts,
			module:     module,
		}
		return mr, mr.init()
	}
}
