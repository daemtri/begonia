package servicemesh

import (
	"flag"
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"golang.org/x/exp/slog"
	"golang.org/x/net/context"
)

const (
	Name = "servicemesh"
)

func init() {
	component.Register[component.Discovery](Name, &DiscoveryBootloader{})
}

type DiscoveryBootloader struct {
	reg Registry
}

func (d *DiscoveryBootloader) Destroy() error {
	//TODO implement me
	panic("implement me")
}

func (d *DiscoveryBootloader) AddFlags(fs *flag.FlagSet) {

}

func (d *DiscoveryBootloader) ValidateFlags() error {
	return validate.Struct(d)
}

func (d *DiscoveryBootloader) Boot(logger *slog.Logger) error {
	d.reg.logger = logger
	return d.reg.Boot()
}

func (d *DiscoveryBootloader) Retrofit() error {
	//TODO implement me
	panic("implement me")
}

func (d *DiscoveryBootloader) Instance() component.Discovery {
	return &d.reg
}

type Registry struct {
	logger *slog.Logger
}

func (r *Registry) Boot() error {
	return nil
}

func (r *Registry) Register(ctx context.Context, service component.ServiceEntry) error {
	r.logger.Info("register", "service", service)
	return nil
}

func (r *Registry) Lookup(ctx context.Context, id, name string) (se *component.ServiceEntry, err error) {
	return &component.ServiceEntry{
		ID:   id,
		Name: name,
		Endpoints: []string{
			fmt.Sprintf("grpc://%s:80", name),
		},
	}, nil
}

func (r *Registry) Browse(ctx context.Context, name string) (*component.Service, error) {
	return &component.Service{
		Entries: []component.ServiceEntry{
			{
				ID:   "",
				Name: name,
				Endpoints: []string{
					fmt.Sprintf("grpc://%s:80", name),
				},
			},
		},
		Configs: []component.ConfigItem{
			{
				Key:   "LoadBalancingConfig",
				Value: "round_robin",
			},
		},
	}, nil
}

func (r *Registry) Watch(ctx context.Context, name string, ch chan<- *component.Service) error {
	s, _ := r.Browse(ctx, name)
	ch <- s
	<-ctx.Done()
	return nil
}
