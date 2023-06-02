package slicemap

import "sort"

type sortAble[T any] struct {
	data     []T
	lessFunc func(i, j T) bool
}

func (sa sortAble[T]) Len() int {
	return len(sa.data)
}

func (sa sortAble[T]) Less(i, j int) bool {
	return sa.lessFunc(sa.data[i], sa.data[j])
}

// Swap swaps the elements with indexes i and j.
func (sa sortAble[T]) Swap(i, j int) {
	sa.data[i], sa.data[j] = sa.data[j], sa.data[i]
}

func Sort[T any](arr []T, less func(left, right T) bool) {
	sort.Sort(sortAble[T]{data: arr, lessFunc: less})
}

type CompareAble interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func SortASC[T CompareAble](l T, r T) bool {
	return l < r
}
