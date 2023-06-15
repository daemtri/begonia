package app

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/resources"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
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
