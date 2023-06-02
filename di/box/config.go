package box

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/flagx"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
)

type ConfigItem = struct {
	Key   string
	Value string
}

// ConfigLoader defined a interface to load config
type ConfigLoader interface {
	// Load load config from config file or other source,
	// setter is a callback function to set config, it is goroutine safe,
	// that means you can call setter immediately,
	// then start a goroutine to watch config change until context done,
	// and call setter again when config changed
	Load(ctx context.Context, setter func([]ConfigItem)) error
}

type configLoaderBuilder struct {
	ConfigLoader `flag:""`
	source       flagx.Source
	name         string
}

func (cb *configLoaderBuilder) ValidateFlags() error {
	return validate.Struct(cb.ConfigLoader)
}

func (cb *configLoaderBuilder) Build(ctx context.Context) (*configLoaderBuilder, error) {
	return cb, nil
}
