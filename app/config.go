package app

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
)

type runtimeConfigLoader struct {
	driver component.Configuration
}

func (c *runtimeConfigLoader) Load(ctx context.Context, setter func([]box.ConfigItem)) error {
	c.driver = box.Invoke[component.Configuration](ctx)
	return nil
}
