package app

import (
	"context"

	"github.com/daemtri/begonia/app/resources"
	"github.com/daemtri/begonia/di/box"
	"github.com/daemtri/begonia/grpcx"
	"github.com/daemtri/begonia/pkg/helper"
	"github.com/daemtri/begonia/runtime/component"
	"google.golang.org/grpc"
)

var (
	servicesConns     helper.OnceMap[string, grpc.ClientConnInterface]
	grpcClientBuilder *grpcx.ClientBuilder
	configWatcher     component.Configurator
	distrubutedLocker component.DistrubutedLocker
	resourcesManager  *resources.Manager
)

func initGlobal(ctx context.Context) error {
	grpcClientBuilder = box.Invoke[*grpcx.ClientBuilder](ctx)
	configWatcher = box.Invoke[component.Configurator](ctx)
	resourcesManager = box.Invoke[*resources.Manager](ctx)
	distrubutedLocker = box.Invoke[component.DistrubutedLocker](ctx)
	return nil
}
