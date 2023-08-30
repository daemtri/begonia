package balancer

import (
	_ "github.com/daemtri/begonia/grpcx/balancer/p2c"
	_ "github.com/daemtri/begonia/grpcx/balancer/ringhash"
	_ "github.com/daemtri/begonia/grpcx/balancer/specify"
	_ "google.golang.org/grpc/balancer/rls"
	_ "google.golang.org/grpc/balancer/roundrobin"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	_ "google.golang.org/grpc/balancer/weightedtarget"
)

// 官方Balancer介绍 https://github.com/grpc/grpc/blob/master/doc/load-balancing.md
// serviceConfig: https://github.com/grpc/grpc/blob/master/doc/service_config.md
