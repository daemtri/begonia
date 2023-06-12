package redis

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/driver/redis"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	mapset "github.com/deckarep/golang-set/v2"
)

var (
	Name = "redis"
)

func init() {
	component.Register[component.DistrubutedLocker](Name, &DistrubutedLockerBootloader{})
}

// DistrubutedLockerBootloader
type DistrubutedLockerBootloader struct {
	DistrubutedLocker

	addr     string
	db       int
	username string
	password string
}

func (d *DistrubutedLockerBootloader) AddFlags(fs *flag.FlagSet) {
	flag.StringVar(&d.addr, "addr", "127.0.0.1:6379", "redis addr")
	flag.IntVar(&d.db, "db", 0, "redis db")
	flag.StringVar(&d.username, "username", "", "redis username")
	flag.StringVar(&d.password, "password", "", "redis password")
	flag.DurationVar(&d.Expiration, "expiration", 10*time.Second, "expiration period")
	flag.DurationVar(&d.Renewal, "renewal", 3*time.Second, "renewal interval")
}

func (d *DistrubutedLockerBootloader) ValidateFlags() error {
	return nil
}

func (d *DistrubutedLockerBootloader) Boot(log *logx.Logger) error {
	d.LockValue = fmt.Sprintf("%s:%s", runtime.GetServiceName(), runtime.GetServiceID())
	d.Logger = log
	var err error
	d.Client, err = redis.NewRedis(context.Background(), &redis.Option{
		Addr:     d.addr,
		DB:       d.db,
		Username: d.username,
		Password: d.password,
	})
	return err
}

func (d *DistrubutedLockerBootloader) Retrofit() error {
	return nil
}

func (d *DistrubutedLockerBootloader) Instance() component.DistrubutedLocker {
	return &d.DistrubutedLocker
}

func (d *DistrubutedLockerBootloader) Destroy() error {
	return d.Client.Close()
}

type DistrubutedLocker struct {
	Client     *redis.Redis
	Expiration time.Duration
	Renewal    time.Duration
	Logger     *logx.Logger
	LockValue  string
}

func (dl *DistrubutedLocker) GetLock(ctx context.Context, name string) component.Locker {
	return NewLocker(dl, ctx, name)
}

type Locker struct {
	*DistrubutedLocker
	ctx                context.Context
	key                string
	lockMap            mapset.Set[string]
	ticker             time.Ticker
	closeToStopRenewal chan struct{}
}

func NewLocker(dl *DistrubutedLocker, ctx context.Context, name string) *Locker {
	return &Locker{
		DistrubutedLocker: dl,
		ctx:               ctx,
		key:               fmt.Sprintf("app:lock:%s", name),
		ticker:            *time.NewTicker(dl.Renewal),
	}
}

var lockerContextKey = &struct{ name string }{name: "redis-locker"}

func (l *Locker) tryLockFromContext() bool {
	value := l.ctx.Value(lockerContextKey)
	if value == nil {
		l.lockMap = mapset.NewSet(Name)
		return false
	}
	lockMap, ok := value.(mapset.Set[string])
	if !ok {
		panic(errors.New("invalid locker context"))
	}
	if !lockMap.Contains(l.key) {
		lockMap.Add(l.key)
		return false
	}
	return true
}

func (l *Locker) renewal() context.Context {
	l.closeToStopRenewal = make(chan struct{})
	go func() {
		defer l.ticker.Stop()
		for {
			select {
			case <-l.ticker.C:
				_, err := l.Client.CmpRefresh(l.ctx, l.key, l.LockValue, l.Expiration)
				if err != nil {
					l.Logger.Error("failed to extend lock", "name", l.key)
				}
			case <-l.closeToStopRenewal:
				return
			}
		}
	}()
	if l.lockMap == nil {
		return l.ctx
	}
	return context.WithValue(l.ctx, lockerContextKey, l.lockMap)
}

func (l *Locker) Lock() (context.Context, error) {
	if l.tryLockFromContext() {
		return l.ctx, nil
	}
	for {
		ok, err := l.Client.SetNX(
			l.ctx,
			l.key,
			l.LockValue,
			l.Expiration,
		).Result()
		if ok {
			break
		}
		l.Logger.Warn("failed to acquire lock", "name", l.key, "error", err)
		select {
		case <-l.ctx.Done():
			return nil, context.DeadlineExceeded
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	return l.renewal(), nil
}

func (l *Locker) TryLock() (context.Context, error) {
	if l.tryLockFromContext() {
		return l.ctx, nil
	}
	ok, err := l.Client.SetNX(
		l.ctx,
		l.key,
		l.LockValue,
		l.Expiration,
	).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return l.ctx, errors.New("acquire lock failed, unknow error")
	}
	return l.renewal(), nil
}

func (l *Locker) Unlock() {
	if l.closeToStopRenewal == nil {
		return
	}
	close(l.closeToStopRenewal)
	_, err := l.Client.CmpDel(l.ctx, l.key, l.LockValue)
	if err != nil {
		l.Logger.Error("failed to release lock", "error", err, "name", l.key)
	}
}
