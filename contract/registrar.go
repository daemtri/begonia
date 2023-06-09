package contract

import "context"

// RouteCell please use app.Route(msgid,handleFunc)
type RouteCell struct {
	MsgID      int32
	HandleFunc func(ctx context.Context, req []byte) error
}

type RouteRegistrar interface {
	RegisterRoute(routes ...RouteCell)
}

// TaskProcessorRegistrar
type TaskProcessorRegistrar interface {
	ProcessTask(taskType string, handle func(context.Context, *Task) error)
}

type PubSubConsumerRegistrar interface {
	SubscribeTopic(topic string, handle any)
}
