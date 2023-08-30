package bootstrap

import (
	"context"
	"fmt"

	"github.com/daemtri/begonia/contract"
)

// RouteRegistrar 路由注册表
type RouteRegistrar struct {
	routes map[int32]func(ctx context.Context, req []byte) error
}

func NewRouteRegistrar() (*RouteRegistrar, error) {
	return &RouteRegistrar{
		routes: make(map[int32]func(ctx context.Context, req []byte) error),
	}, nil
}

func (rr *RouteRegistrar) RegisterRoute(routes ...contract.RouteCell) {
	for _, route := range routes {
		if _, ok := rr.routes[route.MsgID]; ok {
			panic(fmt.Errorf("route %d already registered", route.MsgID))
		}
		rr.routes[route.MsgID] = route.HandleFunc
	}
}
