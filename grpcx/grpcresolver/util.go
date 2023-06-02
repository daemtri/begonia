package grpcresolver

type Metadata map[string]string

func (m Metadata) Equal(o interface{}) bool {
	old, ok := o.(Metadata)
	if !ok {
		return false
	}
	for k := range m {
		v2, ok := old[k]
		if !ok {
			return false
		}
		if v2 != m[k] {
			return false
		}
	}
	return true
}
