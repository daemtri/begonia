package box

import (
	"github.com/daemtri/begonia/di"
)

type options struct {
	opts       []di.Option
	flagPrefix string
}

func newOptions() *options {
	return &options{
		opts: make([]di.Option, 0, 4),
	}
}

type Option interface {
	apply(o *options)
}

type optionsFunc func(o *options)

func (of optionsFunc) apply(o *options) { of(o) }

func WithName(name string) Option {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithName(name))
	})
}

func WithFlags(prefix string) Option {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithFlagset(nfs.FlagSet(prefix)))
		o.flagPrefix = prefix
	})
}

func WithOptional[T any](fn func(name string, err error)) Option {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithOptional[T](fn))
	})
}

// WithSelect 仅供在ProvideInject时使用，可以指定注入某个类型的名字
func WithSelect[T any](name string) Option {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithSelect[T](name))
	})
}

func WithOverride() Option {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithOverride())
	})
}
