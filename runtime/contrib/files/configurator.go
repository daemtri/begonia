package files

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/chanpubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/filepathx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slog"
	"sigs.k8s.io/yaml"
)

const Name = "file"

func init() {
	component.Register[component.Configurator](Name, &ConfiguratorBootloader{})
}

type ConfiguratorBootloader struct {
	dir string

	instance Configurator
}

func (c *ConfiguratorBootloader) Destroy() error {
	return c.instance.close()
}

func (c *ConfiguratorBootloader) AddFlags(fs *flag.FlagSet) {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = filepathx.UserHomePath(".sgr")
	}
	fs.StringVar(&c.dir, "dir", filepath.Join(pwd, "configs"), "配置文件目录")
}

func (c *ConfiguratorBootloader) ValidateFlags() error {
	return validate.Struct(c)
}

func (c *ConfiguratorBootloader) Boot(logger *slog.Logger) error {
	c.instance.log = logger
	path, err := filepath.Abs(c.dir)
	if err != nil {
		return err
	}
	c.instance.configDir = path
	c.instance.broker = chanpubsub.NewBroker[fsnotify.Event]()
	return c.instance.init()
}

func (c *ConfiguratorBootloader) Retrofit() error {
	//TODO implement me
	panic("implement me")
}

func (c *ConfiguratorBootloader) Instance() component.Configurator {
	return &c.instance
}

type Configurator struct {
	log       *slog.Logger
	configDir string

	watcher *fsnotify.Watcher
	broker  *chanpubsub.Broker[fsnotify.Event]
}

func (c *Configurator) close() error {
	c.log.Info("结束监听配置文件", "config dir", c.configDir)
	if err := c.watcher.Close(); err != nil {
		return fmt.Errorf("watcher关闭失败: %w", err)
	}
	return nil
}

func (c *Configurator) init() error {
	c.log.Info("开始监听配置文件", "dir", c.configDir)
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
					c.log.Info("配置文件发生变化", "event", event.String(), "path", event.Name)
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

func (c *Configurator) WatchConfig(ctx context.Context, name string) component.Iterator[component.ConfigDecoder] {
	cfgFile := filepath.Join(c.configDir, name+".yaml")
	return &configfileIterator{Configurator: c, cfgFile: cfgFile, ctx: ctx}
}

func (c *Configurator) ReadConfig(ctx context.Context, name string) (component.ConfigDecoder, error) {
	cfgFile := filepath.Join(c.configDir, name+".yaml")
	return readFile(cfgFile)
}

type configfileIterator struct {
	*Configurator
	cfgFile string
	once    sync.Once
	updates <-chan fsnotify.Event
	cancel  func()
	ctx     context.Context
}

func (c *configfileIterator) Next() (rtn component.ConfigDecoder, err error) {
	c.once.Do(func() {
		rtn, err = readFile(c.cfgFile)
		c.log.Info("添加监听配置文件", "file", c.cfgFile, "path", c.configDir)
		if err := c.watcher.Add(c.cfgFile); err != nil {
			c.log.Warn("监听配置出错", "file", c.cfgFile, "error", err)
		}
		c.updates, c.cancel = c.broker.Subscribe(c.cfgFile)
	})
	if rtn != nil || err != nil {
		return
	}
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case event, ok := <-c.updates:
		if !ok {
			return nil, io.ErrUnexpectedEOF
		}
		c.log.Info("配置文件更新", "event", event.Name)
		rtn, err = readFile(c.cfgFile)
		return
	}
}

func (c *configfileIterator) Stop() {
	c.cancel()
}

func readFile(configFile string) (component.ConfigDecoder, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件出错 path=%s, err=%w", configFile, err)
	}
	return component.NewConfigDecoder(data, func(raw []byte, x any) error {
		return yaml.Unmarshal(raw, x)
	}), nil
}
