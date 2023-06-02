package chanpubsub

import (
	"sync"
	"time"
)

// Broker 处理消息的订阅和取消订阅
type Broker[U any] struct {
	pubMutex   sync.Mutex
	publishers map[string]chan U // 消息发布队列

	subMutex       sync.RWMutex
	subscribers    map[string]map[any]chan U // 消息订阅队列, topic=>channel==>chan
	subscribersCnt map[string]map[any]int    // 消息订阅队列计数
}

// NewBroker 创建一个 broker
func NewBroker[U any]() *Broker[U] {
	b := &Broker[U]{
		publishers:     map[string]chan U{},
		subscribers:    map[string]map[any]chan U{},
		subscribersCnt: map[string]map[any]int{},
	}
	return b
}

// Topic 返回一个Topic的publisher，不如topic不存在，则创建
func (b *Broker[U]) Topic(topic string) chan<- U {
	b.pubMutex.Lock()
	defer b.pubMutex.Unlock()

	if ch, ok := b.publishers[topic]; ok {
		return ch
	}
	publisher := make(chan U, 1)
	b.publishers[topic] = publisher
	go b.serveTopic(topic)
	return publisher
}

func (b *Broker[U]) serveTopic(topic string) {
	updates := b.publishers[topic]
	for update := range updates {
		b.subMutex.RLock()
		for _, subscriber := range b.subscribers[topic] {
			select {
			case subscriber <- update:
			case <-time.After(10 * time.Millisecond):
				// TODO: 10ms超时,打印日志
			}
		}
		b.subMutex.RUnlock()
	}
	// 销毁topic
	b.removeTopic(topic)
}

// Subscribe 订阅一个topic，创建一个channel，并返回
// 同时返回一个cancel函数，用于取消订阅
// 当 channel为空时，则自动生成唯一ID
func (b *Broker[U]) Subscribe(topic string, opts ...SubscribeOpt[U]) (updates <-chan U, cancel func()) {
	b.subMutex.Lock()
	defer b.subMutex.Unlock()

	// 解析参数
	options := subscribeOpts[U]{}
	for i := range opts {
		opts[i](&options)
	}

	var subChan chan U
	defer func() {
		if options.message != nil {
			subChan <- *options.message
		}
	}()

	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[any]chan U)
		b.subscribersCnt[topic] = make(map[any]int)
	}

	channel := options.channel
	// 当channel为空时，默认使用唯一消费者模式
	if channel == "" {
		subChan = make(chan U, 1)
		b.subscribers[topic][subChan] = subChan
		b.subscribersCnt[topic][subChan] = 1
		return subChan, func() {
			b.unsubscribe(topic, subChan)
		}
	}

	// 当channel不为空时，并且channel存在时
	if _, ok := b.subscribers[topic][channel]; ok {
		b.subscribersCnt[topic][channel]++
		subChan = b.subscribers[topic][channel]
		return subChan, func() {
			b.unsubscribe(topic, channel)
		}
	}

	// 当channel不存在时
	subChan = make(chan U, 1)
	b.subscribers[topic][channel] = subChan
	b.subscribersCnt[topic][channel] = 1
	return subChan, func() {
		b.unsubscribe(topic, channel)
	}
}

// Unsubscribe 从指定topic的订阅列表删除删除订阅通道
func (b *Broker[U]) unsubscribe(topic string, channel any) {
	b.subMutex.Lock()
	defer b.subMutex.Unlock()

	if _, ok := b.subscribers[topic]; ok {
		if _, ok := b.subscribers[topic][channel]; ok {
			b.subscribersCnt[topic][channel]--
			if b.subscribersCnt[topic][channel] == 0 {
				close(b.subscribers[topic][channel])
				delete(b.subscribers[topic], channel)
				delete(b.subscribersCnt[topic], channel)
			}
		}
	}
}

// removeTopic 从publishers中删除topic，并关闭所有订阅通道
// 并删除所有订阅者
func (b *Broker[U]) removeTopic(topic string) {
	b.subMutex.Lock()
	defer b.subMutex.Unlock()
	for _, subscriber := range b.subscribers[topic] {
		close(subscriber)
	}
	delete(b.subscribers, topic)
	delete(b.subscribersCnt, topic)

	b.pubMutex.Lock()
	defer b.pubMutex.Unlock()
	delete(b.publishers, topic)
}
