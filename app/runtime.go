package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"

	"git.bianfeng.com/stars/wegame/wan/wanx/app/depency"
	"git.bianfeng.com/stars/wegame/wan/wanx/app/pubsub"
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
		return conn
	})
}

type cluserConnProxy struct {
	id   string
	conn grpc.ClientConnInterface
}

func newClusterConnProxy(id string, conn grpc.ClientConnInterface) grpc.ClientConnInterface {
	return &cluserConnProxy{
		id:   id,
		conn: conn,
	}
}

func (ccp *cluserConnProxy) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	ctx2 := metadata.AppendToOutgoingContext(ctx, "service-id", ccp.id)
	return ccp.conn.Invoke(ctx2, method, args, reply, opts...)
}

func (ccp *cluserConnProxy) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx2 := metadata.AppendToOutgoingContext(ctx, "service-id", ccp.id)
	return ccp.conn.NewStream(ctx2, desc, method, opts...)
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
	return newClusterConnProxy(id, conn)
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

func (mc *moduleConfig[T]) parserConfig(dec component.ConfigDecoder) error {
	newInstance, kind := mc.init()
	if kind == reflect.Pointer {
		if err := dec.Decode(newInstance); err != nil {
			return err
		}
	} else {
		if err := dec.Decode(&newInstance); err != nil {
			return err
		}
	}
	mc.instance.Store(newInstance)
	return nil
}

// SpanWatch 实现contract.ConfigInterface SpanWatch接口
func (mc *moduleConfig[T]) SpanWatch(ctx context.Context, setter func(T) error) {
	cfg, err := configWatcher.ReadConfig(ctx, mc.name)
	if err != nil {
		panic(fmt.Errorf("read config error: %w", err))
	}
	if err := mc.parserConfig(cfg); err != nil {
		panic(fmt.Errorf("parse config error: %w", err))
	}
	if err := setter(mc.Instance()); err != nil {
		panic(fmt.Errorf("set config error: %w", err))
	}
	logger.Info("module config watch start", "name", mc.name)
	iterator := configWatcher.WatchConfig(ctx, mc.name)
	go func() {
		for {
			cfg, err := iterator.Next()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					logger.Info("module config watch timeout", "name", mc.name)
					return
				}
				logger.Error("module config watch error", "name", mc.name, "error", err)
				return
			}
			if err := mc.parserConfig(cfg); err != nil {
				logger.Error("module config watch parse error", "name", mc.name, "error", err)
			}
			if err := setter(mc.Instance()); err != nil {
				logger.Error("module config watch setter error", "name", mc.name, "error", err)
				// TODO: terminate app
			} else {
				logger.Info("module config watch setter success", "name", mc.name)
			}
		}
	}()
}

// GetConfig 获取配置
func GetConfig[T any](ctx context.Context) contract.ConfigInterface[T] {
	mr := objectContainerFromCtx(ctx)
	return mr.config.MustGetOrInit(func() any {
		mc := &moduleConfig[T]{
			name: mr.opts.ConfigName,
			init: helper.ZeroWithKind[T],
		}
		mc.preload(ctx)
		return mc
	}).(*moduleConfig[T])
}
