package app

import (
	"context"
	"errors"
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	config "git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/jsonconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"sigs.k8s.io/yaml"
)

type runtimeConfigLoader struct {
	driver component.Configurator
}

func (c *runtimeConfigLoader) Load(ctx context.Context, setter func([]box.ConfigItem)) error {
	c.driver = box.Invoke[component.Configurator](ctx)

	cfg, err := c.driver.ReadConfig(ctx, appConfigName)
	if err != nil {
		return fmt.Errorf("读取配置文件格式出错, name=%s,err=%w", appConfigName, err)
	}
	jsonRawConfig, err := yaml.YAMLToJSON(cfg.Raw())
	if err != nil {
		return fmt.Errorf("转化配置文件格式出错, name=%s, err=%w", appConfigName, err)
	}
	logger.Info("app config load", "name", appConfigName, "raw", cfg.Raw())
	items, err := config.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return fmt.Errorf("解析配置文件出错, name=%s, err=%w", appConfigName, err)
	}
	setter(items)
	go func() {
		iterator := c.driver.WatchConfig(appConfigName)
		for {
			cfg, err := iterator.Next(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					logger.Info("module config watch timeout", "name", appConfigName)
					return
				}
				panic(err)
			}
			logger.Info("app config update", "name", appConfigName, "raw", cfg.Raw())
			jsonRawConfig, err := yaml.YAMLToJSON(cfg.Raw())
			if err != nil {
				panic(err)
			}
			items, err := config.ParseJSONToKeyValue(string(jsonRawConfig))
			if err != nil {
				panic(err)
			}
			setter(items)
		}
	}()
	return nil
}
