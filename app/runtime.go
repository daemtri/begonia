package app

import (
	"context"
	"database/sql"

	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"github.com/redis/go-redis/v9"
)

type contextKey struct{ name string }

var (
	moduleRuntimeCtxKey = &contextKey{name: "module_runtime"}
)

type moduleOption struct {
	Dependecies []string `flag:"dependecies" usage:"依赖"`
}

type moduleRuntime struct {
	moduleName string
	opts       *moduleOption
	module     Module

	redisClient helper.OnceCell[*redis.Client]
	dbClient    helper.OnceCell[*sql.DB]
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
	mr := moduleRuntimeFromCtx(ctx)
	return mr.redisClient.GetOrInit(nil)
}

// GetDB  获取数据库
func GetDB(ctx context.Context, name string) *sql.DB {
	mr := moduleRuntimeFromCtx(ctx)
	return mr.dbClient.GetOrInit(nil)
}

// GetConfig 获取配置
func GetConfig(ctx context.Context) contract.ConfigInterface {
	panic("unimplement")
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

// GetClient 获取grpc服务，从 ctx中获取名称, 同时支持cluster和service
func GetClient[T any](ctx context.Context) T {
	panic("unimplement")
}

// GetAuth 获取认证信息
func GetAuth(ctx context.Context) contract.AuthInterface {
	panic("unimplement")
}
