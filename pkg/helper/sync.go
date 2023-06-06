package helper

import (
	"sync"

	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/syncx"
)

type OnceCell[T any] struct {
	once sync.Once
	cell T
}

func (oc *OnceCell[T]) GetOrInit(fn func() T) T {
	oc.once.Do(func() {
		oc.cell = fn()
	})
	return oc.cell
}

type OnceMap[K comparable, T any] struct {
	data syncx.Map[K, OnceCell[T]]
}

func (om *OnceMap[K, T]) GetOrInit(key K, fn func() T) T {
	actual, _ := om.data.LoadOrStore(key, OnceCell[T]{})
	return actual.GetOrInit(fn)
}
