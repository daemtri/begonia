package app

import (
	"fmt"
	"net/http"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/pubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/resources"
	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box/config/yamlconfig"
	"git.bianfeng.com/stars/wegame/wan/wanx/driver/kafka"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx/tracing"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"

	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/files"
	_ "git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/k8s"
	_ "git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/nacos"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/redis"
	_ "git.bianfeng.com/stars/wegame/wan/wanx/runtime/contrib/servicemesh"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
)

var (
	broadCastHost     string
	enableSideCarMode bool
	logger            = logx.GetLogger("app")
	appConfigName     string
)

func Run(name string) {
	runtime.SetServiceName(name)

	box.FlagSet().StringVar(&broadCastHost, "broadcast-host", broadCastHost, "默认广播地址")
	box.FlagSet().BoolVar(&enableSideCarMode, "sidecar-enable", false, "开启sgr服务发现边车模式")
	box.FlagSet().StringVar(&appConfigName, "app-config", fmt.Sprintf("app_%s", runtime.GetServiceAlias()), "app配置文件路径")

	// 注册基础功能
	box.Provide[*grpcx.ClientBuilder](grpcx.NewClientBuilder, box.WithFlags("grpc-client"))
	box.Provide[*grpcx.ServerBuilder](grpcx.NewServerBuilder, box.WithFlags("grpc-server"))
	box.Provide[*tracing.Factory](&tracing.Factory{
		ServiceInstanceID: runtime.GetServiceID(),
		ServiceNamespace:  runtime.GetNamespace(),
		ServiceName:       runtime.GetServiceName(),
		ServiceVersion:    runtime.GetServiceVersion(),
	}, box.WithFlags("trace"))

	// 注册runtime
	box.Provide[component.Configurator](&runtime.Builder[component.Configurator]{Name: files.Name}, box.WithFlags("config"))
	box.Provide[component.Discovery](&runtime.Builder[component.Discovery]{Name: files.Name}, box.WithFlags("discovery"))
	box.Provide[component.DistrubutedLocker](&runtime.Builder[component.DistrubutedLocker]{Name: redis.Name}, box.WithFlags("lock"))

	// 注册bootstrap
	box.Provide[*bootstrap.RouteRegistrar](bootstrap.NewRouteRegistrar)
	box.Provide[*bootstrap.ServiceRegistrar](bootstrap.NewServiceRegistrar)
	box.Provide[*bootstrap.ContextInjector](bootstrap.NewContextInjector)
	box.Provide[*bootstrap.BusinessService](bootstrap.NewBusinessService)
	box.Provide[bootstrap.Server](bootstrap.NewLogicServer, box.WithFlags("grpc-server"), box.WithName("grpc"))
	box.Provide[bootstrap.Server](bootstrap.NewHttpServer, box.WithFlags("http-server"), box.WithName("http"))
	box.Provide[bootstrap.Runable](func(server bootstrap.Server) bootstrap.Runable { return server },
		box.WithName("grpc"), box.WithSelect[bootstrap.Server]("grpc"),
	)
	box.Provide[bootstrap.Runable](func(server bootstrap.Server) bootstrap.Runable { return server },
		box.WithName("http"), box.WithSelect[bootstrap.Server]("http"),
	)
	box.Provide[bootstrap.Engine](bootstrap.NewEngine)

	// 注册app相关功能
	box.Provide[GrpcServiceRegistrar](newGrpcServiceRegistrarImpl)
	box.Provide[*kafka.Producer](kafka.NewProducer, box.WithFlags("kafka-producer"))
	box.Provide[pubsub.Publisher](pubsub.NewKafkaPublisher)
	box.Provide[contract.PubSubConsumerRegistrar](&mockPubSubConsumerRegistrar{})
	box.Provide[contract.TaskProcessorRegistrar](&mockTaskProcessorRegistrar{})
	box.Provide[*resources.Manager](resources.NewManager, box.WithFlags("resources"))
	box.Provide[chi.Router](newHttpServerMux)
	box.Provide[http.Handler](func(r chi.Router) http.Handler { return r })
	box.Provide[*Integrator](newIntegrator)
	box.Provide[bootstrap.Runable](NewServiceRegisterRunable, box.WithName("register"))

	// 初始化module和服务注册
	box.UseInit(initModules())

	if err := box.Bootstrap[bootstrap.Engine](
		yamlconfig.Init(),
		box.UseConfigLoader("runtime", &runtimeConfigLoader{}),
	); err != nil {
		logger.Error("engine is stopped", "error", err)
	}
}
