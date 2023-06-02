package slicemap

func MapValues[K comparable, T any](m map[K]T) []T {
	arr := make([]T, 0, len(m))
	for i := range m {
		arr = append(arr, m[i])
	}
	return arr
}

func SliceIsEqual[T CompareAble](left, right []T) bool {
	if len(left) != len(right) {
		return false
	}
	Sort[T](left, SortASC[T])
	Sort[T](right, SortASC[T])
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func MapIsEqual[K, T comparable](left, right map[K]T) bool {
	if len(left) != len(right) {
		return false
	}
	for k := range left {
		v2, ok := right[k]
		if !ok {
			return false
		}
		if v2 != left[k] {
			return false
		}
	}
	return true
}
