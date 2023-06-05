package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/chanpubsub"
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/syncx"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
)

var (
	logger = logx.GetLogger("runtime/agent")
)

type DiscoveryAgent struct {
	component.Discovery

	cache syncx.Map[string, *component.Service]
	queue *chanpubsub.Broker[*component.Service]
	lock  sync.Mutex
}

func NewDiscoveryAgent(d component.Discovery) *DiscoveryAgent {
	return &DiscoveryAgent{
		Discovery: d,
		queue:     chanpubsub.NewBroker[*component.Service](),
	}
}

// Lookup 查询指定id和name的ServiceEntry
func (da *DiscoveryAgent) Lookup(ctx context.Context, name, id string) (component.ServiceEntry, error) {
	ds, ok := da.cache.Load(name)
	if !ok {
		var err error
		ds, err = da.startWatch(name)
		if err != nil {
			return component.ServiceEntry{}, err
		}
	}
	for i := range ds.Entries {
		if ds.Entries[i].ID == id {
			return ds.Entries[i], nil
		}
	}
	return component.ServiceEntry{}, errors.New("not found")
}

// Browse 查询指定name的所有ServiceEntry
func (da *DiscoveryAgent) Browse(ctx context.Context, name string) (*component.Service, error) {
	ds, ok := da.cache.Load(name)
	if !ok {
		var err error
		ds, err = da.startWatch(name)
		if err != nil {
			return nil, err
		}
	}
	logger.Debug("DiscoveryAgent Browse", "name", name, "entries", ds)
	return ds, nil
}

func (da *DiscoveryAgent) startWatch(name string) (*component.Service, error) {
	da.lock.Lock()
	defer da.lock.Unlock()
	ses2, ok := da.cache.Load(name)
	if ok {
		return ses2, nil
	}
	ch := make(chan *component.Service, 1)
	if err := da.Discovery.Watch(context.TODO(), name, ch); err != nil {
		return nil, fmt.Errorf("discovery watch  error %s", err)
	}
	sender := da.queue.Topic(name)
	ses1 := <-ch
	da.cache.Store(name, ses1)
	logger.Debug("DiscoveryAgent service found", "name", name, "entries", ses1)
	go func() {
		for ses := range ch {
			da.cache.Store(name, ses)
			sender <- ses
			logger.Debug("DiscoveryAgent service changed", "name", name, "entries", ses)
		}
	}()
	return ses1, nil
}

func (da *DiscoveryAgent) Watch(ctx context.Context, name string, ch chan<- *component.Service) error {
	ses, err := da.startWatch(name)
	if err != nil {
		return err
	}
	topic := da.queue.Topic(name)
	topic <- ses

	updates, cancel := da.queue.Subscribe(name)
	for {
		select {
		case s := <-updates:
			ch <- s
		case <-ctx.Done():
			cancel()
			return nil
		}
	}
}
