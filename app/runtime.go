package app

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/depency"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/resources"
	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/grpcx"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey struct{ name string }

var (
	moduleRuntimeCtxKey = &contextKey{name: "module_runtime"}

	servicesConns     helper.OnceMap[string, grpc.ClientConnInterface]
	grpcClientBuilder *grpcx.ClientBuilder
	configWatcher     component.Configuration
	resourcesManager  *resources.Manager
)

func initGlobal(ctx context.Context) error {
	grpcClientBuilder = box.Invoke[*grpcx.ClientBuilder](ctx)
	configWatcher = box.Invoke[component.Configuration](ctx)
	resourcesManager = box.Invoke[*resources.Manager](ctx)
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

// GetModuleName 获取当前Module的名称
func GetModuleName(ctx context.Context) string {
	mr := moduleRuntimeFromCtx(ctx)
	return mr.moduleName
}

// GetLogger 获取日志
func GetLogger(ctx context.Context) *logx.Logger {
	mr := moduleRuntimeFromCtx(ctx)
	return logx.GetLogger("module:" + mr.moduleName)
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
	return servicesConns.MustGetOrInit(name, func() grpc.ClientConnInterface {
		conn, err := grpcClientBuilder.NewGrpcClientConn(name, "grpc://", "")
		if err != nil {
			panic(fmt.Errorf("new grpc client error: name=%s,error=%s", name, err))
		}
		return conn
	})
}

type userInfo struct {
	md metadata.MD
}

func (u *userInfo) get(key string) string {
	ret := u.md.Get("tenant_id")
	if len(ret) == 0 {
		panic(fmt.Errorf("no %s in metadata", key))
	}
	return ret[0]
}

func (u *userInfo) GetTenantID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("tenant_id"))))
}

func (u *userInfo) GetUserID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("user_id"))))
}

func (u *userInfo) GetGameID() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("game_id"))))
}

func (u *userInfo) GetSource() string {
	return u.get("source")
}

func (u *userInfo) GetVersion() uint32 {
	return uint32(helper.Must(strconv.Atoi(u.get("version"))))
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context) contract.UserInfoInterface {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		panic(fmt.Errorf("no metadata in context"))
	}
	return &userInfo{md: md}
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

// moduleConfig 模块配置
type moduleConfig[T any] struct {
	name     string
	init     func() (T, reflect.Kind)
	instance atomic.Value
}

func (mc *moduleConfig[T]) preload(ctx context.Context) {
	cfg, err := configWatcher.ReadConfig(ctx, mc.name)
	if err != nil {
		panic(err)
	}
	newInstance, kind := mc.init()
	if kind == reflect.Pointer {
		if err := cfg.Decode(newInstance); err != nil {
			panic(err)
		}
	} else {
		if err := cfg.Decode(&newInstance); err != nil {
			panic(err)
		}
	}

	mc.instance.Store(newInstance)
}

// MustGet 实现contract.ConfigInterface Instance
func (mc *moduleConfig[T]) Instance() T {
	return mc.instance.Load().(T)
}

// SpanWatch 实现contract.ConfigInterface SpanWatch接口
func (mc *moduleConfig[T]) SpanWatch(ctx context.Context, setter func(T)) error {
	ch, err := configWatcher.WatchConfig(ctx, mc.name)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case cfg := <-ch:
				newInstance, kind := mc.init()
				if kind == reflect.Pointer {
					if err := cfg.Decode(newInstance); err != nil {
						panic(err)
					}
				} else {
					if err := cfg.Decode(&newInstance); err != nil {
						panic(err)
					}
				}
				mc.instance.Store(newInstance)
				setter(newInstance)
			case <-ctx.Done():
				logger.Info("module config watch done", "name", mc.name)
			}
		}
	}()
	return nil
}

// GetConfig 获取配置
func GetConfig[T any](ctx context.Context) contract.ConfigInterface[T] {
	mr := moduleRuntimeFromCtx(ctx)
	return mr.config.MustGetOrInit(func() any {
		mc := &moduleConfig[T]{
			name: fmt.Sprintf("module_%s", mr.moduleName),
			init: helper.NewWithKind[T],
		}
		mc.preload(ctx)
		return mc
	}).(*moduleConfig[T])
}
