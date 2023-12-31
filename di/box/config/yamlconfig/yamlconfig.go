package yamlconfig

import (
	"context"
	"fmt"
	"os"

	"github.com/daemtri/begonia/di/box"
	"github.com/daemtri/begonia/di/box/config/jsonconfig"
	"sigs.k8s.io/yaml"
)

func Init() box.BuildOption {
	return box.UseConfigLoader("", &ConfigLoader{})
}

type ConfigLoader struct {
	ConfigFile string `flag:"config" default:"./configs/config.yaml" usage:"配置文件路径"`
}

func (c *ConfigLoader) Load(ctx context.Context, setter func([]box.ConfigItem)) error {
	items, err := Load(c.ConfigFile)
	if err != nil {
		return err
	}
	if items != nil {
		setter(items)
	}
	return nil
}

func Load(configFile string) ([]box.ConfigItem, error) {
	yamlRawConfig, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("配置文件读取失败: %w", err)
	}
	if len(yamlRawConfig) == 0 {
		return nil, nil
	}
	jsonRawConfig, err := yaml.YAMLToJSON(yamlRawConfig)
	if err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}
	items, err := jsonconfig.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}
	return items, nil
}
