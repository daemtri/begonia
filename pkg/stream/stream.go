package stream

import (
	"context"
	"errors"
)

type Stream[T any] interface {
	Next(ctx context.Context) (T, error)
}

type Chan[T any] <-chan T

func (c Chan[T]) Next(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var x T
		return x, ctx.Err()
	case t, ok := <-c:
		if !ok {
			var x T
			return x, errors.New("channel closed")
		}
		return t, nil
	}
}

func One[T any](x T) Stream[T] {
	ch := make(chan T, 1)
	ch <- x
	return Chan[T](ch)
}

type Func[T any] func(ctx context.Context) (T, error)

func (fs Func[T]) Next(ctx context.Context) (T, error) {
	return fs(ctx)
}
