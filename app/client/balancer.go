package client

import (
	"context"
	"strconv"

	"github.com/daemtri/begonia/grpcx/balancer/ringhash"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/balancer/weightedtarget"
)

var (
	defaultBalancerMap = map[uint8]string{}
)

type BalancerSetter interface {
	appBalancer() (appID uint8, balancer string)
}

type BalancerSetFunc func() (appID uint8, balancer string)

func (bsf BalancerSetFunc) appBalancer() (appID uint8, balancer string) {
	return bsf()
}

func WeightedTarget(appID uint8) BalancerSetter {
	return BalancerSetFunc(func() (appID uint8, balancer string) {
		return appID, weightedtarget.Name
	})
}

func WeightedRoundRobin(appID uint8) BalancerSetter {
	return BalancerSetFunc(func() (appID uint8, balancer string) {
		return appID, weightedroundrobin.Name
	})
}

func RoundRobin(appID uint8) BalancerSetter {
	return BalancerSetFunc(func() (appID uint8, balancer string) {
		return appID, roundrobin.Name
	})
}

func RingHash(appID uint8) BalancerSetter {
	return BalancerSetFunc(func() (appID uint8, balancer string) {
		return appID, ringhash.Name
	})
}

func SetRingHashKeyInt64(ctx context.Context, i int64) context.Context {
	return ringhash.SetRequestHash(ctx, strconv.FormatInt(i, 10))
}

func SetDefaultBalancer(bs ...BalancerSetter) {
	for i := range bs {
		appID, balancer := bs[i].appBalancer()
		defaultBalancerMap[appID] = balancer
	}
}

func GetDefaultBalancer(appID uint8) string {
	if def, ok := defaultBalancerMap[appID]; ok {
		return def
	}
	return roundrobin.Name
}
