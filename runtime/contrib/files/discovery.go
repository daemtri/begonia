package files

import (
	"encoding/json"
	"errors"
	"flag"
	"os"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/validate"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"golang.org/x/exp/slog"
	"golang.org/x/net/context"
	"sigs.k8s.io/yaml"
)

func init() {
	component.Register[component.Discovery](Name, &DiscoveryBootloader{})
}

type DiscoveryBootloader struct {
	ServiceFile string

	reg Registry
}

func (d *DiscoveryBootloader) Destroy() error {
	//TODO implement me
	panic("implement me")
}

func (d *DiscoveryBootloader) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&d.ServiceFile, "appsfile", "./configs/apps.yaml", "服务发现前缀")
}

func (d *DiscoveryBootloader) ValidateFlags() error {
	return validate.Struct(d)
}

func (d *DiscoveryBootloader) Boot(logger *slog.Logger) error {
	d.reg.logger = logger
	d.reg.ServiceFile = d.ServiceFile
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
	ServiceFile string
	logger      *slog.Logger
	services    []component.ServiceEntry
}

func (r *Registry) Boot() error {
	serviceFileRawYaml, err := os.ReadFile(r.ServiceFile)
	if err != nil {
		return err
	}
	serviceFileJSONRaw, err := yaml.YAMLToJSON(serviceFileRawYaml)
	if err != nil {
		return err
	}
	var services []component.ServiceEntry
	if err := json.Unmarshal(serviceFileJSONRaw, &services); err != nil {
		return err
	}
	r.services = services
	return nil
}

func (r *Registry) Register(ctx context.Context, service component.ServiceEntry) error {
	r.logger.Info("register", "service", service)
	return nil
}

func (r *Registry) Lookup(ctx context.Context, id, name string) (se *component.ServiceEntry, err error) {
	for i := range r.services {
		if id == r.services[i].ID && name == r.services[i].Name {
			return &r.services[i], nil
		}
	}
	err = errors.New("service not found")
	return
}

func (r *Registry) Browse(ctx context.Context, name string) (*component.Service, error) {
	ses := make([]component.ServiceEntry, 0, 1)
	for i := range r.services {
		if name == r.services[i].Name {
			ses = append(ses, r.services[i])
		}
	}
	return &component.Service{
		Entries: ses,
	}, nil
}

func (r *Registry) Watch(ctx context.Context, name string, ch chan<- *component.Service) error {
	service, err := r.Browse(ctx, name)
	if err != nil {
		return err
	}
	ch <- service
	<-ctx.Done()
	return nil
}
