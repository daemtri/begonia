package app

import (
	"context"
	"database/sql"
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/config"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/depency"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/header"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/pubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/resources"
	"git.bianfeng.com/stars/wegame/wan/wanx/bootstrap/client"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/redis/go-redis/v9"
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

type moduleOption struct {
	Dependecies []string `flag:"dependecies" usage:"依赖"`
	ConfigName  string   `flag:"config" usage:"配置名,默认为module_{module_name}"`
}

type moduleRuntime struct {
	moduleName string
	opts       *moduleOption
	module     Module
	config     helper.OnceCell[any]
}

func (mr *moduleRuntime) init() error {
	depency.SetModuleConfig(mr.moduleName, mr.opts.Dependecies)
	if mr.opts.ConfigName == "" {
		mr.opts.ConfigName = "module_" + mr.moduleName
	}
	return nil
}

func newModuleRuntime(name string, module Module) func(opts *moduleOption) (*moduleRuntime, error) {
	return func(opts *moduleOption) (*moduleRuntime, error) {
		mr := &moduleRuntime{
			moduleName: name,
			opts:       opts,
			module:     module,
		}
		return mr, mr.init()
	}
}

// GetModuleName 获取当前Module的名称
func GetModuleName(ctx context.Context) string {
	mr := objectContainerFromCtx(ctx)
	return mr.moduleName
}

// GetLogger 获取日志
func GetLogger(ctx context.Context) *logx.Logger {
	mr := objectContainerFromCtx(ctx)
	return logx.GetLogger("module:" + mr.moduleName)
}

// GetLocker 获取分布式锁
// 在同一个ctx上获取同一个key的锁，会直接返回同一个锁对象
func GetLocker(ctx context.Context, key string) component.Locker {
	return distrubutedLocker.GetLock(ctx, key)
}

// GetScheduler 获取定时器
func GetScheduler(ctx context.Context) contract.Scheduler {
	panic("unimplement")
}

// GetMsgPublisher 获取消息队列
func GetMsgPublisher(ctx context.Context) pubsub.Publisher {
	panic("unimplement")
}

// GetServiceConn
func GetServiceConn(ctx context.Context, name string) grpc.ClientConnInterface {
	if !depency.Allow(GetModuleName(ctx), "app", name) {
		panic(fmt.Errorf("module %s not allow to call app %s", GetModuleName(ctx), name))
	}
	return servicesConns.MustGetOrInit(name, func() grpc.ClientConnInterface {
		conn, err := grpcClientBuilder.NewGrpcClientConn(name, "grpc://", "")
		if err != nil {
			panic(fmt.Errorf("new grpc client error: name=%s,error=%s", name, err))
		}
		return client.WrapServiceGrpcClientConn(conn)
	})
}

func GetClusterConn(ctx context.Context, name string, id string) grpc.ClientConnInterface {
	if !depency.Allow(GetModuleName(ctx), "app", name) {
		panic(fmt.Errorf("module %s not allow to call app %s", GetModuleName(ctx), name))
	}
	conn := servicesConns.MustGetOrInit(name, func() grpc.ClientConnInterface {
		conn, err := grpcClientBuilder.NewGrpcClientConn(name, "grpc://", id)
		if err != nil {
			panic(fmt.Errorf("new grpc client error: name=%s,error=%s", name, err))
		}
		return conn
	})
	return client.WrapClusterGrpcClientConn(conn, id)
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context) contract.UserInfoInterface {
	return header.GetUserInfoFromIncomingCtx(ctx)
}

// GetDB  获取数据库
func GetDB(ctx context.Context, name string) *sql.DB {
	if !depency.Allow(GetModuleName(ctx), "db", name) {
		panic(fmt.Errorf("module %s not allow to use db %s", GetModuleName(ctx), name))
	}
	db, err := resourcesManager.GetDB(ctx, name)
	if err != nil {
		panic(err)
	}
	return db
}

// GetRedis 获取redis
func GetRedis(ctx context.Context, name string) *redis.Client {
	if !depency.Allow(GetModuleName(ctx), "redis", name) {
		panic(fmt.Errorf("module %s not allow to use redis %s", GetModuleName(ctx), name))
	}
	redis, err := resourcesManager.GetRedis(ctx, name)
	if err != nil {
		panic(err)
	}
	return redis
}

// GetConfig 获取配置
func GetConfig[T any](ctx context.Context) contract.ConfigInterface[T] {
	mr := objectContainerFromCtx(ctx)
	return mr.config.MustGetOrInit(func() any {
		mc := config.NewModuleConfig(configWatcher, mr.opts.ConfigName, helper.ZeroWithKind[T])
		mc.Preload(ctx)
		return mc
	}).(*config.ModuleConfig[T])
}
