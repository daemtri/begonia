package helper

import "sync"

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
