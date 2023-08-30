package k8s

import (
	"flag"
	"fmt"
	"strings"

	"log/slog"

	"context"

	"github.com/daemtri/begonia/di/box/validate"
	"github.com/daemtri/begonia/runtime/component"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	Name = "k8s"
)

func init() {
	component.Register[component.Discovery](Name, &DiscoveryBootloader{})
}

type DiscoveryBootloader struct {
	Namespace string

	reg Registry
}

func (d *DiscoveryBootloader) Destroy() error {
	//TODO implement me
	panic("implement me")
}

func (d *DiscoveryBootloader) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&d.Namespace, "namespace", "default", "k8s服务发现的namespace")
}

func (d *DiscoveryBootloader) ValidateFlags() error {
	return validate.Struct(d)
}

func (d *DiscoveryBootloader) Boot(logger *slog.Logger) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// Create a Kubernetes client using the configuration.
	// see: https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/service-resources/endpoints-v1/
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	d.reg.logger = logger
	d.reg.clientset = clientset
	d.reg.namespace = d.Namespace
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
	namespace string
	clientset *kubernetes.Clientset
	logger    *slog.Logger
	services  []component.ServiceEntry
}

func (r *Registry) Boot() error {
	return nil
}

func (r *Registry) Register(ctx context.Context, service component.ServiceEntry) error {
	r.logger.Info("register", "service", service)
	return nil
}

func (r *Registry) Lookup(ctx context.Context, name, id string) (se *component.ServiceEntry, err error) {
	s, err := r.Browse(ctx, name)
	if err != nil {
		return nil, err
	}
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			return &s.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (r *Registry) Browse(ctx context.Context, name string) (*component.Service, error) {
	endpoints, err := r.clientset.CoreV1().Endpoints(r.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	serviceEntries := map[string]component.ServiceEntry{}
	for i := range endpoints.Subsets {
		for j := range endpoints.Subsets[i].Addresses {
			id := endpoints.Subsets[i].Addresses[j].IP
			if _, ok := serviceEntries[id]; !ok {
				serviceEntries[id] = component.ServiceEntry{
					ID:        id,
					Alias:     endpoints.Annotations["alias"],
					Version:   endpoints.Annotations["version"],
					Endpoints: make([]string, 0, len(endpoints.Subsets[i].Ports)),
				}
			}
			entry := serviceEntries[id]
			for k := range endpoints.Subsets[i].Ports {
				entry.Endpoints = append(entry.Endpoints, fmt.Sprintf("%s+%s://%s:%d",
					endpoints.Subsets[i].Ports[k].Name,
					strings.ToLower(string(endpoints.Subsets[i].Ports[k].Protocol)),
					endpoints.Subsets[i].Addresses[j].IP,
					endpoints.Subsets[i].Ports[k].Port),
				)
			}
			serviceEntries[id] = entry
		}
	}
	s := &component.Service{
		Entries: make([]component.ServiceEntry, 0, len(serviceEntries)),
	}
	for id := range serviceEntries {
		s.Entries = append(s.Entries, serviceEntries[id])
	}
	return s, nil
}

func (r *Registry) Watch(ctx context.Context, name string) component.Stream[*component.Service] {
	w, err := r.clientset.CoreV1().Endpoints(r.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	resultChan := w.ResultChan()
	return component.StreamFunc[*component.Service](func(stop bool) (*component.Service, error) {
		if stop {
			w.Stop()
			return nil, nil
		}
		if err != nil {
			defer func() { err = nil }()
			return nil, err
		}
		select {
		case <-ctx.Done():
			w.Stop()
			return nil, ctx.Err()
		case event := <-resultChan:
			switch event.Type {
			case watch.Added, watch.Modified:
				ep, ok := event.Object.(*corev1.Endpoints)
				fmt.Println(ep, ok)
			case watch.Deleted:
				fmt.Println("delete", event.Object)
			}
		}
		return nil, nil
	})
}
