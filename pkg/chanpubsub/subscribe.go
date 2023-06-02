package chanpubsub

// subscribeOpts 订阅配置
type subscribeOpts[U any] struct {
	channel string
	message *U
}

// SubscribeOpt 订阅配置类型
type SubscribeOpt[U any] func(so *subscribeOpts[U])

// WithSubscribeChannel 设置订阅的渠道
// 如果未设置渠道，则会生成为一的渠道
func WithSubscribeChannel[U any](channel string) SubscribeOpt[U] {
	return func(so *subscribeOpts[U]) {
		so.channel = channel
	}
}

// WithMessage 设置第一次返回的数据
// 该数据不通过topic传递，而是直接写入订阅的channel
func WithMessage[U any](msg U) SubscribeOpt[U] {
	return func(so *subscribeOpts[U]) {
		so.message = &msg
	}
}
