package runtime

import (
	"context"
	"flag"
	"reflect"

	"strings"

	"github.com/daemtri/begonia/logx"
	"github.com/daemtri/begonia/runtime/component"
)

var (
	logger = logx.GetLogger("runtime")
)

type Option interface {
	AddFlags(fs *flag.FlagSet)
	ValidateFlags() error
}

// addDash 为数组每一个元素添加`--`前缀
func addDash(opts []string) []string {
	flagOptions := make([]string, 0, len(opts))
	for i := range opts {
		flagOptions = append(flagOptions, "--"+opts[i])
	}
	return flagOptions
}

type Builder[T component.Interface] struct {
	// Name 为组件注册的名称
	Name string `flag:"name" usage:"驱动名称" json:"name"`
	// Options 为组件的配置参数，item形式为："--a=1 --b=2"
	Options []string `flag:"opt" usage:"驱动参数" json:"opt"`
}

func (b *Builder[T]) ApplyToComponent(cb Option) error {
	fs := flag.NewFlagSet(b.Name, flag.ContinueOnError)
	cb.AddFlags(fs)
	if err := fs.Parse(addDash(b.Options)); err != nil {
		return err
	}
	if err := cb.ValidateFlags(); err != nil {
		return err
	}
	return nil
}

func (b *Builder[T]) Build(ctx context.Context) (x T, err error) {
	bootloader, err := component.GetLoader[T](b.Name)
	if err != nil {
		return x, err
	}
	if err := b.ApplyToComponent(bootloader); err != nil {
		return x, err
	}

	if err := bootloader.Boot(logger.With("component", reflectGetTypeName[T](), "driver_name", b.Name)); err != nil {
		return x, err
	}
	return bootloader.Instance(), nil
}

func reflectGetTypeName[T any]() string {
	var x = reflectType[T]()
	return strings.ToLower(x.Name())
}

func reflectType[K any]() reflect.Type {
	return reflect.TypeOf(new(K)).Elem()
}
