package cmap

import "sync"

type mmap[K comparable, V any] map[K]V

func (mp mmap[K, V]) Get(key K) (V, bool) {
	v, ok := mp[key]
	return v, ok
}

func (mp mmap[K, V]) Set(key K, val V) {
	mp[key] = val
}

func (mp mmap[K, V]) Remove(key K) {
	delete(mp, key)
}

func (mp mmap[K, V]) Range(iter func(key K, val V) bool) {
	for key := range mp {
		if !iter(key, mp[key]) {
			break
		}
	}
}

func (mp mmap[K, V]) Length() int {
	return len(mp)
}

type SafeMap[K comparable, V any] struct {
	mm  mmap[K, V]
	mux sync.RWMutex
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		mm: make(mmap[K, V]),
	}
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.mm.Get(key)
}

func (m *SafeMap[K, V]) Set(key K, val V) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.mm.Set(key, val)
}

func (m *SafeMap[K, V]) Remove(key K) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.mm.Remove(key)
}

func (m *SafeMap[K, V]) Length() int {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.mm.Length()
}

func (m *SafeMap[K, V]) Range(iter func(key K, val V) bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	m.mm.Range(iter)
}

type BatchView[K comparable, V any] interface {
	Get(key K) (V, bool)
	Range(iter func(key K, val V) bool)
	Length() int
}

func (m *SafeMap[K, V]) View(fn func(txn BatchView[K, V])) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	fn(m.mm)
}

type BatchUpdate[K comparable, V any] interface {
	BatchView[K, V]
	Set(key K, val V)
	Remove(key K)
}

func (m *SafeMap[K, V]) Update(fn func(txn BatchUpdate[K, V])) {
	m.mux.Lock()
	defer m.mux.Unlock()

	fn(m.mm)
}

func (m *SafeMap[K, V]) ToMap() map[K]V {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.mm
}
