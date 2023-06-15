package redis

import (
	"context"
	"encoding"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/logx"

	"github.com/redis/go-redis/v9"
)

var logger = logx.GetLogger("/gf/driver/redis")

type Options struct {
	Addr            string        `flag:"addr" default:"127.0.0.1:6379" usage:"Redis host and port, like: 127.0.0.1:6379"`
	Username        string        `flag:"username" default:"" usage:"Redis username"`
	Password        string        `flag:"password" default:"" usage:"Redis password"`
	DB              int           `flag:"db" default:"0" usage:"which db used by client"`
	PoolSize        int           `flag:"pool-size" default:"10" usage:"Maximum number of socket connections"`
	PoolTimeout     time.Duration `flag:"pool-timeout" default:"5s" usage:"Amount of time client waits for connection if all connections"`
	DisableDelayCmd bool          `flag:"disable-delay-cmd" default:"true" usage:"disable command, like: keys, flushall, flushdb, hgetall"`
	DefaultKeyTTL   time.Duration `flag:"default-key-ttl" default:"2160h" usage:"key default ttl 90 days"`
	LargeKeyLen     int           `flag:"large-key-len" default:"10240" usage:"the limits of checkLargeKey function"`
}

// Redis 组件
type Redis struct {
	*redis.Client
	refreshScriptHash    string
	delScriptHash        string
	getDelString         string
	hashSafelyDecrString string
	cmpSetScriptHash     string
}

func NewRedis(_ context.Context, option *Options) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        option.Addr,
		Username:    option.Username,
		Password:    option.Password,
		PoolSize:    option.PoolSize,
		PoolTimeout: option.PoolTimeout,
		DB:          option.DB,
	})
	if option.DisableDelayCmd {
		client.AddHook(disableCmd{largeKeyLen: option.LargeKeyLen})
	}

	return &Redis{Client: client}, nil
}

func FromClient(client *redis.Client) (*Redis, error) {
	return &Redis{Client: client}, nil
}

const hashSafelyDecrScript = `
local value = redis.call("HINCRBY", KEYS[1], KEYS[2], ARGV[1])
if value < 0 then
	redis.call("HINCRBY", KEYS[1], KEYS[2], -ARGV[1])
	return value
else
	return value
end
`

// HashSafelyDecr hash 安全减操作，扣减负数,回滚数据,返回负数
func (redis *Redis) HashSafelyDecr(ctx context.Context, key, field string, increment int64) (result int64, err error) {
	reloaded := false
RedisAction:
	if redis.hashSafelyDecrString == "" {
		reloaded = true
		script, e := redis.ScriptLoad(ctx, hashSafelyDecrScript).Result()
		if e != nil {
			return 0, e
		}

		redis.hashSafelyDecrString = script
	}

	result, err = redis.EvalSha(ctx, redis.hashSafelyDecrString, []string{key, field}, increment).Int64()
	if err != nil {
		if !reloaded {
			redis.hashSafelyDecrString = ""
			goto RedisAction
		}
		return 0, err
	}
	return result, nil
}

const getDelScript = `
	local value = redis.call("GET", KEYS[1])
	if( value ) then
		redis.call("DEL", KEYS[1])
		return value
	else
		return value
	end
`

// GetDel 获取并删除数据
func (redis *Redis) GetDel(ctx context.Context, key string) (result int64, err error) {
	reloaded := false
RedisAction:
	if redis.getDelString == "" {
		reloaded = true
		getDelHash, err := redis.ScriptLoad(ctx, getDelScript).Result()
		if err != nil {
			return 0, err
		}

		redis.getDelString = getDelHash
	}

	result, err = redis.EvalSha(ctx, redis.getDelString, []string{key}).Int64()
	if err != nil {
		if !reloaded {
			redis.getDelString = ""
			goto RedisAction
		}
		return 0, err
	}
	return result, nil
}

const delScript = `
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end`

func (redis *Redis) CmpDel(ctx context.Context, key string, v string) (int, error) {
	reloaded := false
RedisAction:
	if redis.delScriptHash == "" {
		reloaded = true
		scriptHash, err := redis.ScriptLoad(ctx, delScript).Result()
		if err != nil {
			return 0, err
		}
		redis.delScriptHash = scriptHash
	}
	result, err := redis.EvalSha(ctx, redis.delScriptHash, []string{key}, v).Int()
	if err != nil {
		if !reloaded {
			redis.delScriptHash = ""
			goto RedisAction
		}
		return 0, err
	}
	return result, nil
}

const refreshScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1], ARGV[2])
	else
		return 0
	end`

func (redis *Redis) CmpRefresh(ctx context.Context, key string, v string, duration time.Duration) (int, error) {
	reloaded := false
RedisCmpRefreshAction:
	if redis.refreshScriptHash == "" {
		reloaded = true
		scriptHash, err := redis.ScriptLoad(ctx, refreshScript).Result()
		if err != nil {
			return 0, err
		}
		redis.refreshScriptHash = scriptHash
	}
	result, err := redis.EvalSha(ctx, redis.refreshScriptHash, []string{key}, v, strconv.Itoa(int(duration.Milliseconds()))).Int()
	if err != nil {
		if !reloaded {
			redis.refreshScriptHash = ""
			goto RedisCmpRefreshAction
		}
		return 0, err
	}
	return result, nil
}

const cmpSetScript = `
if redis.call("get",KEYS[1]) == ARGV[1] then
	return redis.call("set", KEYS[1], ARGV[2])
else
    return 0
end`

func (redis *Redis) CmpSet(ctx context.Context, key string, oldV string, v string) (int, error) {
	reloaded := false
RedisCmpSetAction:
	if redis.cmpSetScriptHash == "" {
		reloaded = true
		scriptHash, err := redis.ScriptLoad(ctx, cmpSetScript).Result()
		if err != nil {
			return 0, err
		}
		redis.cmpSetScriptHash = scriptHash
	}
	result, err := redis.EvalSha(ctx, redis.cmpSetScriptHash, []string{key}, oldV, v).Int()
	if err != nil {
		if !reloaded {
			redis.cmpSetScriptHash = ""
			goto RedisCmpSetAction
		}
		return 0, err
	}
	return result, nil
}

// disableCmd redis 命令禁用hook
type disableCmd struct {
	largeKeyLen int
}

func isDisableCmd(cmds ...redis.Cmder) error {
	for _, cmd := range cmds {
		switch strings.ToLower(cmd.Name()) {
		case "keys", "flushall", "flushdb":
			return fmt.Errorf("disable cmd: %s", cmd.Name())
		}
	}

	return nil
}

func (th disableCmd) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (th disableCmd) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		checkLargeKey(th.largeKeyLen, cmd)
		if err := isDisableCmd(cmd); err != nil {
			return err
		}
		return next(ctx, cmd)
	}
}

func (th disableCmd) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		checkLargeKey(th.largeKeyLen, cmds...)
		if err := isDisableCmd(cmds...); err != nil {
			return err
		}
		return next(ctx, cmds)
	}
}

func checkLargeKey(largeKeyLen int, cmds ...redis.Cmder) {
	for _, cmd := range cmds {
		args := cmd.Args()
		if len(args) <= 2 {
			continue
		}

		vals := args[2:]
		for _, val := range vals {
			switch v := val.(type) {
			case string:
				s := val.(string)
				size := len(s)
				if size >= largeKeyLen {
					largeKeyLog(args[0].(string), args[1].(string), size)
				}
			case []byte:
				s := val.([]byte)
				size := len(s)
				if size >= largeKeyLen {
					largeKeyLog(args[0].(string), args[1].(string), size)
				}
			case encoding.BinaryMarshaler:
				b, err := v.MarshalBinary()
				if err == nil {
					size := len(b)
					if size >= largeKeyLen {
						largeKeyLog(args[0].(string), args[1].(string), size)
					}
				}
			}
		}
	}
}

func largeKeyLog(command, key string, size int) {
	logger.Warn("checkLargeKey", "msg", "redis key too large.", "command", command, "key", key, "size", size)
}
