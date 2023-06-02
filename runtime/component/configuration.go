package component

import (
	"context"
)

type Configuration interface {
	Interface

	// ReadConfig 获取应用配置
	ReadConfig(ctx context.Context, name string) ([]ConfigItem, error)
	// WatchConfig 获取应用配置
	WatchConfig(ctx context.Context, name string) (<-chan []ConfigItem, error)
}

type ConfigItem = struct {
	Key   string
	Value string
}
