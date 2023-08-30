package app

import (
	"context"

	"github.com/daemtri/begonia/app/depency"
	"github.com/daemtri/begonia/pkg/helper"
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
	Dependencies []string `flag:"dependencies" usage:"依赖"`
	ConfigName   string   `flag:"config" usage:"配置名,默认为{module_name}"`
}

type moduleRuntime struct {
	moduleName string
	opts       *moduleOption
	module     Module
	config     helper.OnceCell[any]
}

func (mr *moduleRuntime) init() error {
	depency.SetModuleConfig(mr.moduleName, mr.opts.Dependencies)
	if mr.opts.ConfigName == "" {
		mr.opts.ConfigName = mr.moduleName
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
