package schedule

import (
	"context"

	"github.com/daemtri/begonia/driver/kafka"
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
