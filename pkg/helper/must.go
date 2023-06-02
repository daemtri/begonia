package helper

func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

type ChanHelper[T any] chan T

func (c ChanHelper[T]) TrySend(x T, err error) error {
	if err != nil {
		return err
	}
	c <- x
	return nil
}

func Chain[T any](x chan T) ChanHelper[T] {
	return ChanHelper[T](x)
}
