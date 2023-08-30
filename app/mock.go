package app

import (
	"context"

	"github.com/daemtri/begonia/contract"
)

type mockPubSubConsumerRegistrar struct {
}

func (*mockPubSubConsumerRegistrar) SubscribeTopic(topic string, handle any) {

}

type mockTaskProcessorRegistrar struct {
}

func (*mockTaskProcessorRegistrar) ProcessTask(taskType string, handle func(context.Context, *contract.Task) error) {

}
