package slicemap

func Contains[T comparable](waits []T, v T) bool {
	for i := range waits {
		if waits[i] == v {
			return true
		}
	}
	return false
}
