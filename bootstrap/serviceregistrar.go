package bootstrap

import (
	"fmt"

	"google.golang.org/grpc"
)

// ServiceRegistrar 服务注册表
type ServiceRegistrar struct {
	services map[*grpc.ServiceDesc]any
}

func NewServiceRegistrar() (*ServiceRegistrar, error) {
	return &ServiceRegistrar{
		services: make(map[*grpc.ServiceDesc]any),
	}, nil
}

func (s *ServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl any) {
	if _, ok := s.services[desc]; ok {
		panic(fmt.Errorf("service %s already registered", desc.ServiceName))
	}
	s.services[desc] = impl
}

func (s *ServiceRegistrar) RegisterTo(sr grpc.ServiceRegistrar) {
	for desc := range s.services {
		sr.RegisterService(desc, s.services[desc])
	}
}
