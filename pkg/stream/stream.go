package stream

type Stream[T any] interface {
	Close()
	Next() (T, error)
}

type onceStream[T any] struct {
	item T
}

func (os *onceStream[T]) Close() {}

func (os *onceStream[T]) Next() (T, error) {
	return os.item, nil
}

func Once[T any](x T) Stream[T] {
	return &onceStream[T]{item: x}
}

type chanStream[T any] struct {
	ch     <-chan T
	cancel func()
}

func (cs *chanStream[T]) Close() {
	cs.cancel()
}

func (cs *chanStream[T]) Next() (T, error) {
	return <-cs.ch, nil
}

func Chan[T any](ch <-chan T, cancel func()) Stream[T] {
	return &chanStream[T]{ch: ch, cancel: cancel}
}

type funcStream[T any] struct {
	next func() (T, error)
	stop func()
}

func (fs *funcStream[T]) Close() {
	fs.stop()
}

func (fs *funcStream[T]) Next() (T, error) {
	return fs.next()
}

func Func[T any](f func() (T, error), stop func()) Stream[T] {
	return &funcStream[T]{next: f, stop: stop}
}
