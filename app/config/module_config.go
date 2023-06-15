package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
)

// moduleConfig 模块配置
type ModuleConfig[T any] struct {
	driver   component.Configurator
	name     string
	init     func() (T, reflect.Kind)
	instance atomic.Value
}

func NewModuleConfig[T any](driver component.Configurator, name string, init func() (T, reflect.Kind)) *ModuleConfig[T] {
	return &ModuleConfig[T]{
		driver: driver,
		name:   name,
		init:   init,
	}
}

func (mc *ModuleConfig[T]) Preload(ctx context.Context) {
	cfg, err := mc.driver.ReadConfig(ctx, mc.name)
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
func (mc *ModuleConfig[T]) Instance() T {
	return mc.instance.Load().(T)
}

func (mc *ModuleConfig[T]) parserConfig(dec component.ConfigDecoder) error {
	newInstance, kind := mc.init()
	if kind == reflect.Pointer {
		if err := dec.Decode(newInstance); err != nil {
			return err
		}
	} else {
		if err := dec.Decode(&newInstance); err != nil {
			return err
		}
	}
	mc.instance.Store(newInstance)
	return nil
}

// SpanWatch 实现contract.ConfigInterface SpanWatch接口
func (mc *ModuleConfig[T]) SpanWatch(ctx context.Context, setter func(T) error) {
	cfg, err := mc.driver.ReadConfig(ctx, mc.name)
	if err != nil {
		panic(fmt.Errorf("read config error: %w", err))
	}
	if err := mc.parserConfig(cfg); err != nil {
		panic(fmt.Errorf("parse config error: %w", err))
	}
	if err := setter(mc.Instance()); err != nil {
		panic(fmt.Errorf("set config error: %w", err))
	}
	logger.Info("module config watch start", "name", mc.name)
	iterator := mc.driver.WatchConfig(ctx, mc.name)
	go func() {
		for {
			cfg, err := iterator.Next()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					logger.Info("module config watch timeout", "name", mc.name)
					return
				}
				logger.Error("module config watch error", "name", mc.name, "error", err)
				return
			}
			if err := mc.parserConfig(cfg); err != nil {
				logger.Error("module config watch parse error", "name", mc.name, "error", err)
			}
			if err := setter(mc.Instance()); err != nil {
				logger.Error("module config watch setter error", "name", mc.name, "error", err)
				// TODO: terminate app
			} else {
				logger.Info("module config watch setter success", "name", mc.name)
			}
		}
	}()
}
