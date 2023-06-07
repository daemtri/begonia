package app

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
)

// moduleConfig 模块配置
type moduleConfig[T any] struct {
	name     string
	once     sync.Once
	init     func() (T, reflect.Kind)
	instance atomic.Value
}

func (mc *moduleConfig[T]) preload(ctx context.Context) {
	cfg, err := configWatcher.ReadConfig(ctx, mc.name)
	if err != nil {
		panic(err)
	}
	newInstance, kind := mc.init()
	if kind == reflect.Pointer {
		if err := cfg.Decode(newInstance); err != nil {
			panic(err)
		}
	} else {
		if err := cfg.Decode(&newInstance); err != nil {
			panic(err)
		}
	}

	mc.instance.Store(newInstance)
}

// MustGet 实现contract.ConfigInterface Instance
func (mc *moduleConfig[T]) Instance() T {
	return mc.instance.Load().(T)
}

// Watch 实现contract.ConfigInterface Watch接口
func (mc *moduleConfig[T]) Watch(ctx context.Context, setter func(T)) error {
	ch, err := configWatcher.WatchConfig(ctx, mc.name)
	if err != nil {
		panic(err)
	}
	for cfg := range ch {
		newInstance, kind := mc.init()
		if kind == reflect.Pointer {
			if err := cfg.Decode(newInstance); err != nil {
				panic(err)
			}
		} else {
			if err := cfg.Decode(&newInstance); err != nil {
				panic(err)
			}
		}
		mc.instance.Store(newInstance)
		setter(newInstance)
	}
	return nil
}

// GetConfig 获取配置
func GetConfig[T any](ctx context.Context) contract.ConfigInterface[T] {
	mr := moduleRuntimeFromCtx(ctx)
	return mr.config.GetOrInit(func() any {
		mc := &moduleConfig[T]{
			name: fmt.Sprintf("module_%s", mr.moduleName),
			init: helper.NewWithKind[T],
		}
		mc.preload(ctx)
		return mc
	}).(*moduleConfig[T])
}
