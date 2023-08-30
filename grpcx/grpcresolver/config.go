package grpcresolver

import "github.com/daemtri/begonia/runtime/component"

type ServiceConfig struct {
	LoadBalancingConfig string
	GrayReleaseConfig   string
}

func parseServiceConfig(scs []component.ConfigItem) (sc ServiceConfig) {
	for i := range scs {
		switch scs[i].Key {
		case "LoadBalancingConfig":
			sc.LoadBalancingConfig = scs[i].Value
		case "GrayReleaseConfig":
			sc.GrayReleaseConfig = scs[i].Value
		}
	}
	return
}
