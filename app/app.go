package app

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/yamlconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/tracing"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"

	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/files"
	_ "git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/k8s"
)

var (
	broadCastHost     string
	enableSideCarMode bool
	logger            = logx.GetLogger("app")
)

func Run(name string) {
	runtime.SetServiceName(name)

	box.FlagSet().StringVar(&broadCastHost, "broadcast-host", broadCastHost, "默认广播地址")
	box.FlagSet().BoolVar(&enableSideCarMode, "sidecar-enable", false, "开启sgr服务发现边车模式")

	// 注册基础功能
	box.Provide[*grpcx.ServerBuilder](grpcx.NewServerBuilder, box.WithFlags("grpc"))
	box.Provide[*tracing.Factory](&tracing.Factory{
		ServiceInstanceID: runtime.GetServiceID(),
		ServiceNamespace:  runtime.GetNamespace(),
		ServiceName:       runtime.GetServiceName(),
		ServiceVersion:    runtime.GetServiceVersion(),
	}, box.WithFlags("trace"))

	// 注册runtime
	box.Provide[component.Configuration](&runtime.Builder[component.Configuration]{Name: files.Name}, box.WithFlags("config"))
	box.Provide[component.Discovery](&runtime.Builder[component.Discovery]{Name: files.Name}, box.WithFlags("discovery"))
	box.Provide[component.Concurrency](&runtime.Builder[component.Concurrency]{Name: "file"}, box.WithFlags("concurrency"))

	// 注册bootstrap
	box.Provide[*bootstrap.RouteRegistrar](bootstrap.NewRouteRegistrar)
	box.Provide[*bootstrap.ServiceRegistrar](bootstrap.NewServiceRegistrar)
	box.Provide[*bootstrap.ContextInjector](bootstrap.NewContextInjector)
	box.Provide[*bootstrap.BusinessService](bootstrap.NewBusinessService)
	box.Provide[bootstrap.Server](bootstrap.NewLogicServer, box.WithFlags("server"))
	box.Provide[bootstrap.Engine](bootstrap.NewEngine)

	// 注册app相关功能
	box.Provide[LogicServiceRegistrar](newLogicServiceRegistrarImpl)
	box.Provide[contract.PubSubConsumerRegistrar](&mockPubSubConsumerRegistrar{})
	box.Provide[contract.TaskProcessorRegistrar](&mockTaskProcessorRegistrar{})
	box.Provide[Registry](newRegistry)

	// 初始化module和服务注册
	box.UseInit(initModules())
	box.UseInit(initRegisterApp)

	if err := box.Bootstrap[bootstrap.Engine](
		yamlconfig.Init(),
		box.UseConfigLoader("runtime", &runtimeConfigLoader{}),
	); err != nil {
		logger.Error("engine is stopped", "error", err)
	}
}
