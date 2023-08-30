package app

import (
	"net/http"

	"github.com/daemtri/begonia/app/config"
	"github.com/daemtri/begonia/app/pubsub"
	"github.com/daemtri/begonia/app/resources"
	"github.com/daemtri/begonia/bootstrap"
	"github.com/daemtri/begonia/contract"
	"github.com/daemtri/begonia/di/box"
	"github.com/daemtri/begonia/di/box/config/yamlconfig"
	"github.com/daemtri/begonia/driver/kafka"
	"github.com/daemtri/begonia/grpcx"
	"github.com/daemtri/begonia/grpcx/tracing"
	"github.com/daemtri/begonia/logx"
	"github.com/daemtri/begonia/runtime"
	"github.com/daemtri/begonia/runtime/component"

	"github.com/daemtri/begonia/runtime/contrib/files"
	_ "github.com/daemtri/begonia/runtime/contrib/k8s"
	_ "github.com/daemtri/begonia/runtime/contrib/nacos"
	"github.com/daemtri/begonia/runtime/contrib/redis"
	_ "github.com/daemtri/begonia/runtime/contrib/servicemesh"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
)

var (
	broadCastHost     string
	enableSideCarMode bool
	logger            = logx.GetLogger("app")
	remoteConfigName  string
)

func Run(name string) {
	runtime.SetServiceName(name)

	box.FlagSet().StringVar(&broadCastHost, "broadcast-host", broadCastHost, "默认广播地址")
	box.FlagSet().BoolVar(&enableSideCarMode, "sidecar-enable", false, "开启sgr服务发现边车模式")
	box.FlagSet().StringVar(&remoteConfigName, "remote-config", "", "远程配置文件路径")

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
	box.Provide[component.Discovery](&runtime.Builder[component.Discovery]{Name: configDiscoveryName}, box.WithFlags("discovery"))
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
		box.UseConfigLoader("app", config.NewAppConfigLoader(remoteConfigName)),
	); err != nil {
		logger.Error("engine is stopped", "error", err)
	}
}
