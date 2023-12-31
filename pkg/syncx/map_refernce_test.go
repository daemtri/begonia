package syncx

import "sync"

// This file contains reference map implementations for unit-tests.

// mapInterface is the interface Map implements.
type mapInterface[K comparable, V any] interface {
	Load(K) (V, bool)
	Store(key K, value V)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	LoadAndDelete(key K) (value V, loaded bool)
	Delete(K)
	Range(func(key K, value V) (shouldContinue bool))
}

// WrapperMap is an implementation of mapInterface that wraps sync.Map
type WrapperMap[K comparable, V any] struct {
	m sync.Map
}

func (m *WrapperMap[K, V]) Load(key K) (value V, ok bool) {
	vany, ok := m.m.Load(key)
	if !ok {
		return
	}
	value = vany.(V)
	return
}

func (m *WrapperMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *WrapperMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	actualany, loaded := m.m.LoadOrStore(key, value)
	actual = actualany.(V)
	return
}

func (m *WrapperMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	vany, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return
	}
	value = vany.(V)
	return
}

func (m *WrapperMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *WrapperMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(kany, vany any) bool {
		key := kany.(K)
		value := vany.(V)
		return f(key, value)
	})
}
