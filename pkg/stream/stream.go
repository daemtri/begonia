package stream

type Stream[T any] interface {
	Next() (T, error)
}

type onetream[T any] struct{ item T }

func (os *onetream[T]) Next() (T, error) {
	return os.item, nil
}

func One[T any](x T) Stream[T] {
	return &onetream[T]{item: x}
}

type Chan[T any] <-chan T

func (c Chan[T]) Next() (T, error) {
	return <-c, nil
}

type Func[T any] func() (T, error)

func (fs Func[T]) Next() (T, error) {
	return fs()
}
