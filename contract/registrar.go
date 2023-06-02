package contract

import "context"

type RouteRegistrar interface {
	RegisterRoute(msgID int32, handleFunc func(ctx context.Context, req []byte) error)
}

// TaskProcessorRegistrar
type TaskProcessorRegistrar interface {
	ProcessTask(taskType string, handle func(context.Context, *Task) error)
}

type PubSubConsumerRegistrar interface {
	SubscribeTopic(topic string, handle any)
}
