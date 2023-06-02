package box

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func init() {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("godotenv.Load failed", "err", err)
		}
	}
}

func UseInit(fn func(context.Context) error, opts ...Option) {
	name := ""
	// 获取当前函数的调用栈信息，skip = 0 表示获取当前函数信息
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		// 获取函数名
		fnName := runtime.FuncForPC(pc).Name()
		// 输出函数信息、文件名、行号
		name = fmt.Sprintf("registerd in function %s at file %s, line %d\n", fnName, file, line)
	}
	opts = append(opts, WithName(name))
	Provide[*initializerFunc](func() (*initializerFunc, error) {
		return &initializerFunc{
			name: name,
			fn:   fn,
		}, nil
	}, opts...)
}

// InitFunc 初始化函数
type initFunc func(context.Context) error

type initializerFunc struct {
	name string
	fn   initFunc
}

type initializer[T any] struct {
	instance T
}

func (it *initializer[T]) Build(ctx context.Context) (*initializer[T], error) {
	// register config loader
	configLoaders := Invoke[All[*configLoaderBuilder]](ctx)

	// parser args and envronment
	printConfig := nfs.FlagSet().Bool("print-config", false, "print configuration information")
	nfs.BindFlagSet(flag.CommandLine, envPrefix)

	// load config from config file or other source
	for i := range configLoaders {
		if err := configLoaders[i].Load(ctx, func(items []ConfigItem) {
			SetConfig(items, configLoaders[i].source)
		}); err != nil {
			return nil, fmt.Errorf("load configuration %s failed: %w", configLoaders[i].source, err)
		}
	}

	// print config
	if *printConfig {
		err := EncodeFlags(os.Stdout)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "EncodeFlags error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	beforeFuncs := Invoke[All[*initializerFunc]](ctx)
	for i := range beforeFuncs {
		if err := beforeFuncs[i].fn(ctx); err != nil {
			return nil, fmt.Errorf("execution initFunc %s returned error: %w", beforeFuncs[i].name, err)
		}
	}

	it.instance = Invoke[T](ctx)
	return it, nil
}
