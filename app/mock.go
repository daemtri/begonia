package app

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
)

type mockPubSubConsumerRegistrar struct {
}

func (*mockPubSubConsumerRegistrar) SubscribeTopic(topic string, handle any) {

}

type mockTaskProcessorRegistrar struct {
}

func (*mockTaskProcessorRegistrar) ProcessTask(taskType string, handle func(context.Context, *contract.Task) error) {

}
