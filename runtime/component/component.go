package component

import (
	"flag"
	"fmt"
	"reflect"
	"sync"

	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
)

type Interface interface{}

// Bootloader 定义了一个组件必须实现的引导启动接口
type Bootloader[T Interface] interface {
	AddFlags(fs *flag.FlagSet)
	ValidateFlags() error
	Boot(logger *logx.Logger) error
	Retrofit() error
	Instance() T
	Destroy() error
}

func Register[T Interface](name string, bl Bootloader[T]) {
	typ := reflectType[T]()
	if err := reg.save(typ, name, bl); err != nil {
		panic(err)
	}
}

func GetLoader[T Interface](name string) (Bootloader[T], error) {
	typ := reflectType[T]()
	v, err := reg.load(typ, name)
	if err != nil {
		return nil, err
	}
	return v.(Bootloader[T]), nil
}

var (
	reg = newContainer()
)

func reflectType[K any]() reflect.Type {
	return reflect.TypeOf(new(K)).Elem()
}

type container struct {
	mm  map[reflect.Type]map[string]any
	mux sync.RWMutex
}

func newContainer() *container {
	return &container{
		mm: make(map[reflect.Type]map[string]any),
	}
}

func (c *container) save(typ reflect.Type, name string, v any) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	m, ok := c.mm[typ]
	if !ok {
		m = make(map[string]any)
		c.mm[typ] = m
	}
	if _, ok := m[name]; ok {
		return fmt.Errorf("%s:%s 已存在", typ, name)
	}
	m[name] = v
	return nil
}

func (c *container) load(typ reflect.Type, name string) (any, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	m, ok := c.mm[typ]
	if !ok {
		return nil, fmt.Errorf("%s未注册任何驱动", typ)
	}
	v, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("%s:%s驱动不存在", typ, name)
	}

	return v, nil
}

type Iterator[T any] interface {
	Stop()
	Next() (T, error)
}
