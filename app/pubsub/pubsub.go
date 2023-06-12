package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/driver/kafka"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, msg any) error
}

type kafkaPublisher struct {
	producer *kafka.Producer
}

func NewKafkaPublisher(producer *kafka.Producer) Publisher {
	p := &kafkaPublisher{
		producer: producer,
	}
	return p
}

func (ks *kafkaPublisher) Publish(ctx context.Context, topic string, msg any) error {
	v, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ks.producer.WriteMessage(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%d:%d", time.Now().UnixNano(), crc32.ChecksumIEEE(v))),
		Value: v,
	})
}
