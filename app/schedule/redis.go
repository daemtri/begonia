package schedule

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/driver/kafka"
)

type kafkaScheduler struct {
	producer *kafka.Producer
}

func NewKafkaScheduler(producer *kafka.Producer) Scheduler {
	p := &kafkaScheduler{
		producer: producer,
	}
	return p
}

func (ks *kafkaScheduler) AddTask(ctx context.Context, typename string, task any, opts ...TaskOption) error {
	return nil
}
