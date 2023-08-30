package redis

// import (
// 	"fmt"
// 	"github.com/daemtri/begonia/app"
// 	"github.com/daemtri/begonia/pkg/goroutine"
// 	"github.com/go-redsync/redsync/v4"
// 	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
// 	"os"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )

// type Concurrency struct {
// 	opts        *ConcurrencyOptions
// 	base        *redsync.Redsync
// 	valuePrefix string

// 	mus    map[string]*Mutex
// 	musReq map[string]int
// 	musmu  sync.Mutex
// }

// func (c *Concurrency) getMutex(name string) *Mutex {
// 	if _, ok := c.mus[name]; !ok {
// 		opts := []redsync.Option{
// 			redsync.WithGenValueFunc(func() (string, error) {
// 				return fmt.Sprintf("%s,goroutine=%d",
// 					c.valuePrefix,
// 					goroutine.ID(),
// 				), nil
// 			}),
// 			redsync.WithExpiry(c.opts.Expiry),
// 			redsync.WithTries(c.opts.Tries),
// 		}
// 		c.mus[name] = &Mutex{
// 			cc:       c,
// 			name:     name,
// 			redisMux: c.base.NewMutex(name, opts...),
// 		}
// 	}
// 	return c.mus[name]
// }

// func (c *Concurrency) freeMutex(name string) {
// 	c.musmu.Lock()
// 	defer c.musmu.Unlock()
// 	c.musReq[name]--
// 	if c.musReq[name] <= 0 {
// 		delete(c.mus, name)
// 		delete(c.musReq, name)
// 	}
// }

// func (c *Concurrency) Mutex(name string) *Mutex {
// 	c.musmu.Lock()
// 	defer c.musmu.Unlock()
// 	c.musReq[name]++
// 	return c.getMutex(name)
// }

// type Mutex struct {
// 	name string

// 	cc             *Concurrency
// 	recursion      int32
// 	owner          int64
// 	localMux       sync.Mutex
// 	redisMux       *redsync.Mutex
// 	redisMuxExtend time.Duration
// 	stopChan       chan struct{}
// }

// func (m *Mutex) Lock() error {
// 	gid := goroutine.ID()
// 	if atomic.LoadInt64(&m.owner) == gid {
// 		m.recursion++
// 		return nil
// 	}
// 	// 先锁进程
// 	m.localMux.Lock()
// 	atomic.StoreInt64(&m.owner, gid)
// 	m.recursion = 1
// 	// 再锁Redis
// 	if err := m.redisMux.Lock(); err != nil {
// 		return err
// 	}
// 	go func() {
// 		for {
// 			select {
// 			case <-time.After(m.redisMuxExtend):
// 				ok, err := m.redisMux.Extend()
// 				if !ok || err != nil {
// 					panic(fmt.Errorf("续约失败,key=%s,ok=%t,err=%s", m.name, ok, err))
// 				}
// 			case <-m.stopChan:
// 				return
// 			}
// 		}
// 	}()
// 	return nil
// }

// func (m *Mutex) Unlock() (bool, error) {
// 	if atomic.LoadInt64(&m.owner) != goroutine.ID() {
// 		return false, fmt.Errorf("invalid onwer(%d) to unlock onwer(%d)`s lock", goroutine.ID(), m.owner)
// 	}
// 	// 清理锁
// 	defer m.cc.freeMutex(m.name)
// 	m.recursion--
// 	if m.recursion != 0 {
// 		m.stopChan <- struct{}{}
// 		return true, nil
// 	}
// 	atomic.StoreInt64(&m.owner, -1)
// 	m.localMux.Unlock()
// 	return m.redisMux.Unlock()
// }

// type ConcurrencyOptions struct {
// 	Tries  int           `flag:"tries" default:"32" usage:"最大尝试次数"`
// 	Expiry time.Duration `flag:"expiry" default:"60s" usage:"过期时间"`
// 	Extend time.Duration `flag:"extend" default:"1s" usage:"续约周期"`
// }

// func NewConcurrency(opt *ConcurrencyOptions, client *Redis) (*Concurrency, error) {
// 	pool := goredis.NewPool(client.Client)
// 	c := &Concurrency{
// 		opts:   opt,
// 		base:   redsync.New(pool),
// 		mus:    map[string]*Mutex{},
// 		musReq: map[string]int{},
// 	}
// 	hostname, err := os.Hostname()
// 	if err != nil {
// 		return nil, err
// 	}
// 	c.valuePrefix = fmt.Sprintf(
// 		"app=%s,instance=%s,hostname=%s,proccess=%d",
// 		app.GetServiceName(),
// 		app.GetServiceID(),
// 		hostname,
// 		os.Getpid(),
// 	)
// 	return c, nil
// }
