package di

type Default[T any] interface {
	Default() T
}
