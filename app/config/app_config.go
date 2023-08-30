package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/daemtri/begonia/di/box"
	config "github.com/daemtri/begonia/di/box/config/jsonconfig"
	"github.com/daemtri/begonia/runtime/component"
	"sigs.k8s.io/yaml"
)

type AppConfigLoader struct {
	appConfigName string
	driver        component.Configurator
}

func NewAppConfigLoader(appConfigName string) *AppConfigLoader {
	return &AppConfigLoader{
		appConfigName: appConfigName,
	}
}

func (c *AppConfigLoader) Load(ctx context.Context, setter func([]box.ConfigItem)) error {
	c.driver = box.Invoke[component.Configurator](ctx)

	cfg, err := c.driver.ReadConfig(ctx, c.appConfigName)
	if err != nil {
		return fmt.Errorf("读取配置文件格式出错, name=%s,err=%w", c.appConfigName, err)
	}
	jsonRawConfig, err := yaml.YAMLToJSON(cfg.Raw())
	if err != nil {
		return fmt.Errorf("转化配置文件格式出错, name=%s, err=%w", c.appConfigName, err)
	}
	logger.Info("app config load", "name", c.appConfigName, "raw", cfg.Raw())
	items, err := config.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return fmt.Errorf("解析配置文件出错, name=%s, err=%w", c.appConfigName, err)
	}
	setter(items)
	go func() {
		iterator := c.driver.WatchConfig(ctx, c.appConfigName)
		for {
			cfg, err := iterator.Next()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					logger.Info("module config watch timeout", "name", c.appConfigName)
					return
				}
				panic(err)
			}
			logger.Info("app config update", "name", c.appConfigName, "raw", cfg.Raw())
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
