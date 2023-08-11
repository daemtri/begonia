package component

import (
	"flag"

	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
)

// Bootloader 定义了一个组件必须实现的引导启动接口
type Bootloader[T any] interface {
	AddFlags(fs *flag.FlagSet)
	ValidateFlags() error
	Boot(logger *logx.Logger) error
	Retrofit() error
	Instance() T
	Destroy() error
}
