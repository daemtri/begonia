package grpcresolver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const (
	DefaultEndpointSchema            = "grpc://"
	DefaultUpdateEmptyConnStateDelay = 5 * time.Minute
)

var (
	services = map[string][]component.ServiceEntry{}
)

func SetServiceInLocal(entries ...component.ServiceEntry) {
	for i := range entries {
		name := entries[i].Name
		_, ok := services[name]
		if ok {
			services[name] = append(services[name], entries[i])
		} else {
			services[name] = []component.ServiceEntry{entries[i]}
		}
	}
}

// sgrResolver 封装component.ServiceRegistrar实现了"google.golang.org/grpc/resolver.Resolver"
type sgrResolver struct {
	target     *targetInfo
	clientConn resolver.ClientConn

	disableServiceConfig bool

	discovery component.Discovery

	ctx       context.Context
	ctxCancel context.CancelFunc

	currentServiceEntries []component.ServiceEntry
	currentServiceConfig  ServiceConfig

	schema string
}

func (sr *sgrResolver) Init() error {
	sr.schema = DefaultEndpointSchema
	if schema := sr.target.target.URL.Query().Get("schema"); schema != "" {
		sr.schema = schema
	}
	if ses, ok := services[sr.target.serviceName]; ok {
		sr.currentServiceEntries = ses
		sr.updateClientConnState()
		return nil
	}

	sr.ctx, sr.ctxCancel = context.WithCancel(context.Background())
	iter := sr.discovery.Watch(sr.ctx, sr.target.serviceName)
	service, err := iter.Next()
	if err != nil {
		return err
	}
	// TODO: 增加超时机制
	sr.currentServiceEntries = service.Entries
	sr.currentServiceConfig = parseServiceConfig(service.Configs)
	sr.updateClientConnState()

	go sr.watch(iter)
	logger.Info("Sgr GRPC resolver 初始化成功", "schema", sr.schema)
	return nil
}

func (sr *sgrResolver) updateClientConnState() {
	sis := sr.currentServiceEntries
	sc := sr.currentServiceConfig

	logger.Info("正在变更本地服务发现信息", sis, sc)
	if sc.GrayReleaseConfig != "" {
		sis = grayReleaseHandle(sis, sc.GrayReleaseConfig)
	}
	address := make([]resolver.Address, 0, len(sis))
	for i := range sis {
		md := attributes.New("id", sis[i].ID)
		md = md.WithValue("name", sis[i].Name)
		for k, v := range sis[i].Metadata {
			md = md.WithValue(k, v)
		}
		endpoint := ""
		for _, current := range sis[i].Endpoints {
			if strings.HasPrefix(current, sr.schema) {
				endpoint = strings.TrimPrefix(current, sr.schema)
			}
		}
		if endpoint == "" && len(sis[i].Endpoints) == 1 {
			endpoint = sis[i].Endpoints[0]
		}
		addr := resolver.Address{
			Addr:               endpoint,
			ServerName:         sr.target.serviceName,
			Attributes:         md,
			BalancerAttributes: md,
		}
		address = append(address, addr)
	}
	state := resolver.State{
		Addresses:  address,
		Attributes: attributes.New("resolver", "sgr"),
	}
	if !sr.disableServiceConfig && sc.LoadBalancingConfig != "" {
		// TODO: 负载均衡配置
		// see https://github.com/grpc/grpc/blob/master/doc/service_config.md
		serviceConfigJSON := fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, sc.LoadBalancingConfig)
		state.ServiceConfig = sr.clientConn.ParseServiceConfig(serviceConfigJSON)
	}
	logger.Info("服务状态已更新", "state", state)

	err := sr.clientConn.UpdateState(state)
	if err != nil && len(sis) != 0 {
		logger.Error("更新GRPC客户端链接失败", err)
	}
}

func (sr *sgrResolver) watch(iter component.Iterator[*component.Service]) {
	emptyTimer := time.NewTimer(DefaultUpdateEmptyConnStateDelay)
	emptyTimer.Stop()
	ch := make(chan *component.Service, 1)
	go func() {
		for {
			service, err := iter.Next()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					close(ch)
					return
				}
				logger.Warning("服务发现迭代出错", "err", err)
				return
			}
			ch <- service
		}
	}()
	for {
		select {
		case service, ok := <-ch:
			if !ok {
				logger.Infof("服务解析Watch服务发现组件通道已关闭")
				return
			}
			logger.Infoln("sgrResolver watch update", "service", service)
			sr.currentServiceEntries = service.Entries
			sr.currentServiceConfig = parseServiceConfig(service.Configs)
			if len(service.Entries) == 0 {
				emptyTimer.Reset(DefaultUpdateEmptyConnStateDelay)
			} else {
				sr.updateClientConnState()
			}
		case <-emptyTimer.C:
			if len(sr.currentServiceEntries) == 0 {
				sr.updateClientConnState()
			}
		case <-sr.ctx.Done():
			logger.Warning("sgrResolver watch已取消")
			return
		}
	}
}

// ResolveNow 会被gRPC调用来尝试解析target name
// 可能会被同时并发调用或者多次调用
func (sr *sgrResolver) ResolveNow(opt resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (sr *sgrResolver) Close() {
	sr.ctxCancel()
	sr.ctx.Deadline()
}
