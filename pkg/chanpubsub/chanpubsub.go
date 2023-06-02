package chanpubsub

func Topic[T any](broker Broker[any], name string) chan<- T {
	ret := make(chan T, 1)
	ch := broker.Topic(name)
	go func() {
		for v := range ret {
			ch <- v
		}
	}()
	return ret
}

func Subscribe[T any](broker Broker[any], topic string, opts ...SubscribeOpt[any]) (<-chan T, func()) {
	ret := make(chan T, 1)
	ch, cancel := broker.Subscribe(topic, opts...)
	go func() {
		for v := range ch {
			ret <- v.(T)
		}
	}()
	return ret, cancel
}
