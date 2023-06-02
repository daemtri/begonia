package bootstrap

import (
	"context"
	"fmt"
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

func (rr *RouteRegistrar) RegisterRoute(msgID int32, handleFunc func(ctx context.Context, req []byte) error) {
	if _, ok := rr.routes[msgID]; ok {
		panic(fmt.Errorf("route %d already registered", msgID))
	}
	rr.routes[msgID] = handleFunc
}
