package servicemesh

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"time"

	"log/slog"

	"context"

	"github.com/daemtri/begonia/di/box/validate"
	"github.com/daemtri/begonia/runtime"
	"github.com/daemtri/begonia/runtime/component"
	"github.com/redis/go-redis/v9"
)

const (
	Name = "servicemesh"
)

func init() {
	component.Register[component.Discovery](Name, &DiscoveryBootloader{})
}

type DiscoveryBootloader struct {
	reg           Registry
	redisAddr     string
	redisUsername string
	redisPassword string
	redisDB       int
	PodIP         string
}

func (d *DiscoveryBootloader) Destroy() error {
	//TODO implement me
	panic("implement me")
}

func (d *DiscoveryBootloader) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&d.redisAddr, "redis-addr", "127.0.0.1:6379", "redis address")
	fs.StringVar(&d.redisUsername, "redis-username", "", "redis username")
	fs.StringVar(&d.redisPassword, "redis-password", "", "redis password")
	fs.IntVar(&d.redisDB, "redis-db", 0, "redis db")
	fs.StringVar(&d.PodIP, "pod_ip", "", "pod ip")
}

func (d *DiscoveryBootloader) ValidateFlags() error {
	return validate.Struct(d)
}

func (d *DiscoveryBootloader) Boot(logger *slog.Logger) error {
	d.reg.logger = logger
	d.reg.redisClient = redis.NewClient(&redis.Options{
		Addr:     d.redisAddr,
		Username: d.redisUsername,
		Password: d.redisPassword,
		DB:       d.redisDB,
	})
	d.reg.podIP = d.PodIP
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
	logger      *slog.Logger
	redisClient *redis.Client
	podIP       string
}

func (r *Registry) Boot() error {
	return nil
}

func registerKey(name, id string) string {
	return fmt.Sprintf("wan:%s:apps:%s:%s", runtime.GetNamespace(), name, id)
}

func (r *Registry) Register(ctx context.Context, service component.ServiceEntry) error {
	r.logger.Info("register", "service", service)
	se, err := json.Marshal(service)
	if err != nil {
		return err
	}
	ret := r.redisClient.SetNX(ctx, registerKey(service.Name, service.ID), se, 10*time.Second)
	if ret.Err() != nil {
		return ret.Err()
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				r.redisClient.Expire(ctx, registerKey(service.Name, service.ID), 10*time.Second)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (r *Registry) Lookup(ctx context.Context, name, id string) (se *component.ServiceEntry, err error) {
	st := runtime.ParseServiceType(name)
	switch st {
	case runtime.ServiceTypeCluster:
		ret := r.redisClient.Get(ctx, registerKey(name, id))
		if ret.Err() != nil {
			return nil, ret.Err()
		}
		se := component.ServiceEntry{}
		err = json.Unmarshal([]byte(ret.Val()), &se)
		if err != nil {
			return nil, err
		}
		return &se, nil
	case runtime.ServiceTypeService:
		return &component.ServiceEntry{
			ID:   id,
			Name: name,
			Endpoints: []string{
				fmt.Sprintf("grpc://%s:80", name),
			},
		}, nil
	default:
		panic(fmt.Errorf("unknown service type %s", st))
	}
}

func (r *Registry) Browse(ctx context.Context, name string) (*component.Service, error) {
	st := runtime.ParseServiceType(name)
	switch st {
	case runtime.ServiceTypeCluster:
		ret := r.redisClient.Keys(ctx, registerKey(name, "*"))
		if ret.Err() != nil {
			return nil, ret.Err()
		}
		ids := ret.Val()
		s := &component.Service{
			Entries: make([]component.ServiceEntry, 0, len(ids)),
		}
		for _, id := range ids {
			se := component.ServiceEntry{}
			err := json.Unmarshal([]byte(id), &se)
			if err != nil {
				return nil, err
			}
			s.Entries = append(s.Entries, se)
			s.Configs = []component.ConfigItem{
				{
					Key:   "LoadBalancingConfig",
					Value: "specify",
				},
			}
		}
		return s, nil
	default:
		return &component.Service{
			Entries: []component.ServiceEntry{
				{
					ID:   "unknown",
					Name: name,
					Endpoints: []string{
						fmt.Sprintf("grpc://%s:80", name),
					},
				},
			},
			Configs: []component.ConfigItem{
				{
					Key:   "LoadBalancingConfig",
					Value: "pick_first",
				},
			},
		}, nil
	}
}

func (r *Registry) Watch(ctx context.Context, name string) component.Stream[*component.Service] {
	var lastServcie *component.Service
	return component.StreamFunc[*component.Service](func(stop bool) (*component.Service, error) {
		if stop {
			return nil, fmt.Errorf("stop watch")
		}
		for {
			if lastServcie != nil {
				time.Sleep(1 * time.Second)
			}
			s, err := r.Browse(ctx, name)
			if err != nil {
				return nil, err
			}
			if lastServcie == nil {
				lastServcie = s
				return s, nil
			}
			if !reflect.DeepEqual(lastServcie, s) {
				lastServcie = s
				return s, nil
			}
		}
	})
}
