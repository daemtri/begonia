package kafka

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Consumer 消息消费组件
type Consumer struct {
	opts *ConsumerOption
}

type ConsumerOption struct {
	Brokers string `flag:"brokers" default:"127.0.0.1:9092" usage:"Kafka bootstrap Brokers to connect to, as a comma separated list"`
	Group   string `flag:"group" usage:"消费组"`
}

func NewConsumer(opt *ConsumerOption) (*Consumer, error) {
	return &Consumer{opts: opt}, nil
}

type Reader = kafka.Reader

func (th *Consumer) NewReader(ctx context.Context, topics ...string) *Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(th.opts.Brokers, ","), // Kafka brokers
		GroupID:     th.opts.Group,
		GroupTopics: topics,
	})
}
