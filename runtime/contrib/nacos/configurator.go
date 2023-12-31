package nacos

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"log/slog"

	"github.com/daemtri/begonia/di/box/validate"
	"github.com/daemtri/begonia/runtime/component"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sigs.k8s.io/yaml"
)

const Name = "nacos"

func init() {
	component.Register[component.Configurator](Name, &ConfiguratorBootloader{})
}

type ConfiguratorBootloader struct {
	constant.ClientConfig
	instance    Configurator
	serverAddrs string
}

func (c *ConfiguratorBootloader) Destroy() error {
	return c.instance.close()
}

func (c *ConfiguratorBootloader) AddFlags(fs *flag.FlagSet) {
	fs.Uint64Var(&c.TimeoutMs, "timeout_ms", 10000, "timeout for requesting Nacos server, default value is")
	fs.StringVar(&c.NamespaceId, "namespace_id", "", "the namespaceId of Nacos")
	fs.StringVar(&c.Endpoint, "endpoint", "", "the endpoint for ACM. https://help.aliyun.com/document_detail/130146.html")
	fs.StringVar(&c.RegionId, "region_id", "", "the regionId for ACM & KMS")
	fs.StringVar(&c.AccessKey, "access_key", "", "the accessKey for ACM & KMS")
	fs.StringVar(&c.SecretKey, "secret_key", "", "the secretKey for ACM & KMS")
	fs.BoolVar(&c.OpenKMS, "open_kms", false, `it's to open KMS, default is false. https://help.aliyun.com/product/28933.html, to enable encrypt/decrypt, DataId should be start with "cipher-"`)
	fs.StringVar(&c.CacheDir, "cache_dir", "", "the directory for persist nacos service info,default value is current path")
	fs.StringVar(&c.Username, "username", "", "the username for nacos auth")
	fs.StringVar(&c.Password, "password", "", "the password for nacos auth")
	fs.StringVar(&c.LogDir, "log_dir", "", "the directory for log, default is current path")
	fs.StringVar(&c.serverAddrs, "server_addrs", "", "the server address for nacos")
}

func (c *ConfiguratorBootloader) ValidateFlags() error {
	return validate.Struct(c)
}

func (c *ConfiguratorBootloader) Boot(logger *slog.Logger) error {
	c.instance.log = logger
	var sc []constant.ServerConfig
	serverAddrs := strings.Split(c.serverAddrs, ",")
	for _, addr := range serverAddrs {
		ipPort := strings.SplitN(addr, ":", 2)
		if len(ipPort) != 2 {
			panic(fmt.Errorf("invalid server address %s", addr))
		}
		port, err := strconv.Atoi(ipPort[1])
		if err != nil {
			panic(fmt.Errorf("invalid server address %s", addr))
		}
		sc = append(sc, *constant.NewServerConfig(ipPort[0], uint64(port)))
	}
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &c.ClientConfig,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return err
	}
	c.instance.client = client
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
	log    *slog.Logger
	client config_client.IConfigClient
}

func (c *Configurator) close() error {
	// return c.client.CancelListenConfig(vo.ConfigParam{})
	return nil
}

func (c *Configurator) init() error {
	return nil
}

func parseNameAndGroup(name string) (string, string) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 1 {
		return parts[0], "global"
	}
	return parts[1], parts[0]
}

func (c *Configurator) WatchConfig(ctx context.Context, name string) component.Stream[component.ConfigDecoder] {
	group, parsedName := parseNameAndGroup(name)
	return newWatcher(c, parsedName, group, ctx)
}

func (c *Configurator) ReadConfig(ctx context.Context, name string) (component.ConfigDecoder, error) {
	group, parsedName := parseNameAndGroup(name)
	content, err := c.client.GetConfig(vo.ConfigParam{
		DataId: parsedName,
		Group:  group,
	})
	if err != nil {
		return nil, err
	}
	return yamlConfigFromString(content), nil
}

type yamlConfig struct {
	content string
}

func yamlConfigFromString(s string) *yamlConfig {
	return &yamlConfig{content: s}
}

func (c *yamlConfig) Raw() []byte {
	return []byte(c.content)
}

func (c *yamlConfig) Decode(x any) error {
	return yaml.Unmarshal([]byte(c.content), x)
}

type Watcher struct {
	*Configurator
	dataID   string
	group    string
	contents chan string
	once     sync.Once
	ctx      context.Context
}

func newWatcher(cfg *Configurator, dataID string, group string, ctx context.Context) *Watcher {
	w := &Watcher{
		Configurator: cfg,
		dataID:       dataID,
		group:        group,
		contents:     make(chan string, 100),
		ctx:          ctx,
	}
	return w
}

func (w *Watcher) Next() (component.ConfigDecoder, error) {
	var err error
	w.once.Do(func() {
		err = w.client.ListenConfig(vo.ConfigParam{
			DataId: w.dataID,
			Group:  w.group,
			OnChange: func(namespace, group, dataId, data string) {
				fmt.Println(namespace, group, dataId, data, w.dataID, w.group)
				if dataId == w.dataID && group == w.group {
					w.contents <- data
				}
			},
		})
	})
	if err != nil {
		return nil, err
	}
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case content, ok := <-w.contents:
		if !ok {
			return nil, io.ErrUnexpectedEOF
		}
		return yamlConfigFromString(content), nil
	}
}

func (w *Watcher) Close() error {
	err := w.client.CancelListenConfig(vo.ConfigParam{
		DataId: w.dataID,
		Group:  w.group,
	})
	close(w.contents)
	return err
}

func (w *Watcher) Stop() {
	err := w.Close()
	if err != nil {
		w.log.Warn("close watcher error", "error  ", err)
	}
}
