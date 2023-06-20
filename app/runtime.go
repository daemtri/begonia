package app

import (
	"context"
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/client"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/config"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/depency"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/header"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/pubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/driver/db"
	"git.bianfeng.com/stars/wegame/wan/wanx/driver/redis"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/constraintx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"google.golang.org/grpc"
)

// GeCurrentModule 获取当前Module的名称
func GeCurrentModule(ctx context.Context) string {
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
func GetMsgPublisher(ctx context.Context, name string) pubsub.Publisher {
	if !depency.Allow(GeCurrentModule(ctx), "kafka", name) {
		panic(fmt.Errorf("module %s not allow to use kafka %s", GeCurrentModule(ctx), name))
	}
	p, err := resourcesManager.GetMsgPublisher(ctx, name)
	if err != nil {
		panic(err)
	}
	return p
}

// GetMsgPublisher 获取消息队列
func GetMsgSubscriber(ctx context.Context, name string) pubsub.Subscriber {
	if !depency.Allow(GeCurrentModule(ctx), "kafka", name) {
		panic(fmt.Errorf("module %s not allow to use kafka %s", GeCurrentModule(ctx), name))
	}
	s, err := resourcesManager.GetMsgSubscriber(ctx, name)
	if err != nil {
		panic(err)
	}
	return s
}

// GetServiceConn
func GetService(ctx context.Context, name string) grpc.ClientConnInterface {
	if !depency.Allow(GeCurrentModule(ctx), "app", name) {
		panic(fmt.Errorf("module %s not allow to call app %s", GeCurrentModule(ctx), name))
	}
	return servicesConns.MustGetOrInit(name, func() grpc.ClientConnInterface {
		conn, err := grpcClientBuilder.NewGrpcClientConn(name, "grpc://", "")
		if err != nil {
			panic(fmt.Errorf("new grpc client error: name=%s,error=%s", name, err))
		}
		return client.WrapServiceGrpcClientConn(conn)
	})
}

func GetCluster(ctx context.Context, name string, id string) grpc.ClientConnInterface {
	if !depency.Allow(GeCurrentModule(ctx), "app", name) {
		panic(fmt.Errorf("module %s not allow to call app %s", GeCurrentModule(ctx), name))
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
func GetDB(ctx context.Context, name string) *db.Database {
	if !depency.Allow(GeCurrentModule(ctx), "db", name) {
		panic(fmt.Errorf("module %s not allow to use db %s", GeCurrentModule(ctx), name))
	}
	db, err := resourcesManager.GetDB(ctx, name)
	if err != nil {
		panic(err)
	}
	return db
}

// GetRedis 获取redis
func GetRedis(ctx context.Context, name string) *redis.Redis {
	if !depency.Allow(GeCurrentModule(ctx), "redis", name) {
		panic(fmt.Errorf("module %s not allow to use redis %s", GeCurrentModule(ctx), name))
	}
	redis, err := resourcesManager.GetRedis(ctx, name)
	if err != nil {
		panic(err)
	}
	return redis
}

// GetConfig 获取配置
func GetConfig[T constraintx.Default[T]](ctx context.Context) contract.ConfigInterface[T] {
	mr := objectContainerFromCtx(ctx)
	return mr.config.MustGetOrInit(func() any {
		mc := config.NewModuleConfig[T](configWatcher, mr.opts.ConfigName)
		mc.Preload(ctx)
		return mc
	}).(*config.ModuleConfig[T])
}
