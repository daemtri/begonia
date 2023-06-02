package slicemap

func MapKeys[K comparable, V any](x map[K]V) []K {
	ks := make([]K, 0, len(x))
	for k := range x {
		ks = append(ks, k)
	}
	return ks
}
