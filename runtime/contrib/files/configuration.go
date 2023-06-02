package files

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	config "git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/jsonconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/chanpubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/filepathx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slog"
)

const ConfigName = "file"

func init() {
	component.Register[component.Configuration](ConfigName, &ConfigurationBootloader{})
}

type ConfigurationBootloader struct {
	dir string

	instance Configuration
}

func (c *ConfigurationBootloader) Destroy() error {
	return c.instance.close()
}

func (c *ConfigurationBootloader) AddFlags(fs *flag.FlagSet) {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = filepathx.UserHomePath(".sgr")
	}
	fs.StringVar(&c.dir, "dir", pwd, "配置文件目录")
}

func (c *ConfigurationBootloader) ValidateFlags() error {
	return validate.Struct(c)
}

func (c *ConfigurationBootloader) Boot(logger *slog.Logger) error {
	c.instance.log = logger
	path, err := filepath.Abs(c.dir)
	if err != nil {
		return err
	}
	c.instance.configDir = path
	c.instance.sgrdConfigFile = "sgrd.yaml"
	c.instance.serviceConfigDir = "service_config"
	c.instance.appConfigDir = "app"
	c.instance.broker = chanpubsub.NewBroker[fsnotify.Event]()
	return c.instance.init()
}

func (c *ConfigurationBootloader) Retrofit() error {
	//TODO implement me
	panic("implement me")
}

func (c *ConfigurationBootloader) Instance() component.Configuration {
	return &c.instance
}

type Configuration struct {
	log              *slog.Logger
	configDir        string
	sgrdConfigFile   string
	appConfigDir     string
	serviceConfigDir string

	watcher *fsnotify.Watcher
	broker  *chanpubsub.Broker[fsnotify.Event]
}

func (c *Configuration) close() error {
	c.log.Info("结束监听sgr配置文件", "file", filepath.Join(c.configDir, c.sgrdConfigFile))
	if err := c.watcher.Close(); err != nil {
		return fmt.Errorf("watcher关闭失败: %w", err)
	}
	return nil
}

func (c *Configuration) init() error {
	c.log.Info("开始监听sgr配置文件", "dir", c.configDir)
	var err error
	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event, ok := <-c.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					c.broker.Topic(event.Name) <- event
					c.log.Info("sgr配置文件发生变化", "event", event.String(), "path", event.Name)
				}
			case err, ok := <-c.watcher.Errors:
				if !ok {
					return
				}
				c.log.Warn("watcher监听发生错误", "error", err)
			}
		}
	}()
	return nil
}

func (c *Configuration) readFile(configFile string) ([]component.ConfigItem, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件出错 path=%s, err=%w", configFile, err)
	}
	jsonRawConfig, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("转化配置文件格式出错, path=%s, err=%w", configFile, err)
	}
	items, err := config.ParseJSONToKeyValue(string(jsonRawConfig))
	if err != nil {
		return nil, fmt.Errorf("解析配置文件出错, path=%s, err=%w", configFile, err)
	}
	return items, nil
}

func (c *Configuration) watchConfig(ctx context.Context, configFile string) (<-chan []component.ConfigItem, error) {
	ch := make(chan []component.ConfigItem, 1)
	if err := helper.Chain(ch).TrySend(c.readFile(configFile)); err != nil {
		c.log.Warn("首次读取配置出错", "file", configFile, "error", err)
	}
	if err := c.watcher.Add(configFile); err != nil {
		c.log.Warn("监听配置出错", "file", configFile, "error", err)
	}
	updates, cancel := c.broker.Subscribe(configFile)
	go func() {
		for {
			select {
			case <-ctx.Done():
				cancel()
				return
			case event, ok := <-updates:
				if !ok {
					return
				}
				c.log.Info("配置文件更新", "event", event)
				if err := helper.Chain(ch).TrySend(c.readFile(configFile)); err != nil {
					c.log.Warn("配置文件更新出错", "event", event, "error", err)
				}
			}
		}
	}()
	return ch, nil
}

func (c *Configuration) WatchConfig(ctx context.Context, name string) (<-chan []component.ConfigItem, error) {
	cfgFile := filepath.Join(c.configDir, c.appConfigDir, name+".yaml")
	return c.watchConfig(ctx, cfgFile)
}

func (c *Configuration) ReadConfig(ctx context.Context, name string) ([]component.ConfigItem, error) {
	cfgFile := filepath.Join(c.configDir, c.sgrdConfigFile)
	return c.readFile(cfgFile)
}
