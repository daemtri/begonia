package app

import (
	"errors"
	"flag"

	"log/slog"

	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
)

var (
	configDiscoveryName = "config"
)

func init() {
	component.Register[component.Discovery](configDiscoveryName, &DiscoveryBootloader{})
}

type DiscoveryBootloader struct {
	reg Registry
}

func (d *DiscoveryBootloader) Destroy() error {
	return nil
}

func (d *DiscoveryBootloader) AddFlags(fs *flag.FlagSet) {
	flag.StringVar(&d.reg.path, "path", "apps", "config path")
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
	path   string
}

func (r *Registry) Boot() error {
	return nil
}

func (r *Registry) Register(ctx context.Context, service component.ServiceEntry) error {
	r.logger.Info("register", "service", service)
	return nil
}

func (r *Registry) Lookup(ctx context.Context, name, id string) (se *component.ServiceEntry, err error) {
	service, err := r.Browse(ctx, name)
	if err != nil {
		return nil, err
	}
	for i := range service.Entries {
		if service.Entries[i].ID == id {
			return &service.Entries[i], nil
		}
	}
	return nil, errors.New("not found")
}

func parserService(dec component.ConfigDecoder, name string) (*component.Service, error) {
	var cfgs []component.ServiceEntry
	if err := dec.Decode(&cfgs); err != nil {
		return nil, err
	}
	servce := &component.Service{
		Entries: []component.ServiceEntry{},
	}
	for i := range cfgs {
		if cfgs[i].Name == name {
			servce.Entries = append(servce.Entries, cfgs[i])
		}
	}
	return servce, nil
}

func (r *Registry) Browse(ctx context.Context, name string) (*component.Service, error) {
	dec, err := configWatcher.ReadConfig(ctx, r.path)
	if err != nil {
		return nil, err
	}
	return parserService(dec, name)
}

func (r *Registry) Watch(ctx context.Context, name string) component.Stream[*component.Service] {
	iterator := configWatcher.WatchConfig(ctx, r.path)
	return &configDiscoveryIterator{iter: iterator, name: name}
}

type configDiscoveryIterator struct {
	iter component.Stream[component.ConfigDecoder]
	name string
}

func (cdi *configDiscoveryIterator) Stop() {
	cdi.iter.Stop()
}

func (cdi *configDiscoveryIterator) Next() (*component.Service, error) {
	dec, err := cdi.iter.Next()
	if err != nil {
		return nil, err
	}
	return parserService(dec, cdi.name)
}
