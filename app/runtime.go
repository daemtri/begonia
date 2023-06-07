package app

import (
	"context"
	"database/sql"
	"fmt"

	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type contextKey struct{ name string }

var (
	moduleRuntimeCtxKey = &contextKey{name: "module_runtime"}
	global              = struct {
		redisClient  helper.OnceMap[string, *redis.Client]
		dbClient     helper.OnceMap[string, *sql.DB]
		servicesConn helper.OnceMap[string, grpc.ClientConnInterface]
	}{}
	grpcClientBuilder *grpcx.ClientBuilder
	configWatcher     component.Configuration
)

func initGlobal(ctx context.Context) error {
	grpcClientBuilder = box.Invoke[*grpcx.ClientBuilder](ctx)
	configWatcher = box.Invoke[component.Configuration](ctx)
	return nil
}

type moduleOption struct {
	Dependecies []string `flag:"dependecies" usage:"依赖"`
	ConfigName  string   `flag:"config" usage:"配置名,默认为module_{module_name}.yaml"`
}

type moduleRuntime struct {
	moduleName string
	opts       *moduleOption
	module     Module
	config     helper.OnceCell[any]
}

func withModuleRuntime(ctx context.Context, mr *moduleRuntime) context.Context {
	return context.WithValue(ctx, moduleRuntimeCtxKey, mr)
}

func moduleRuntimeFromCtx(ctx context.Context) *moduleRuntime {
	v := ctx.Value(moduleRuntimeCtxKey)
	if v == nil {
		panic("no module runtime in context")
	}
	return v.(*moduleRuntime)
}

func newModuleRuntime(name string, module Module) func(opts *moduleOption) (*moduleRuntime, error) {
	return func(opts *moduleOption) (*moduleRuntime, error) {
		return &moduleRuntime{
			moduleName: name,
			opts:       opts,
			module:     module,
		}, nil
	}
}

// GetLogger 获取日志
func GetLogger(ctx context.Context) *logx.Logger {
	mr := moduleRuntimeFromCtx(ctx)
	return logx.GetLogger("module:" + mr.moduleName)
}

// GetRedis 获取redis
func GetRedis(ctx context.Context, name string) *redis.Client {
	// mr := moduleRuntimeFromCtx(ctx)
	return global.redisClient.GetOrInit(name, func() *redis.Client {
		return nil
	})
}

// GetDB  获取数据库
func GetDB(ctx context.Context, name string) *sql.DB {
	// mr := moduleRuntimeFromCtx(ctx)
	return global.dbClient.GetOrInit(name, func() *sql.DB {
		return nil
	})
}

// GetLocker 获取分布式锁
func GetLocker(ctx context.Context) contract.DistrubutedLocker {
	panic("unimplement")
}

// GetScheduler 获取定时器
func GetScheduler(ctx context.Context) contract.Scheduler {
	panic("unimplement")
}

// GetPubSub 获取消息队列
func GetPubSub(ctx context.Context) contract.PubSubInterface {
	panic("unimplement")
}

// GetServiceConn
func GetServiceConn(ctx context.Context, name string) grpc.ClientConnInterface {
	// mr := moduleRuntimeFromCtx(ctx)
	return global.servicesConn.GetOrInit(name, func() grpc.ClientConnInterface {
		conn, err := grpcClientBuilder.NewGrpcClientConn(name, "grpc://", "")
		if err != nil {
			panic(fmt.Errorf("new grpc client error: name=%s,error=%s", name, err))
		}
		return conn
	})
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context) contract.UserInfoInterface {
	panic("unimplement")
}
