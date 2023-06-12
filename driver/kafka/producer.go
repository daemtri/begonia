package kafka

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

type ProducerOption struct {
	Brokers string `flag:"brokers" default:"127.0.0.1:9092" usage:"kafka brokers"`
}

// Producer 消息生产组件
type Producer struct {
	opts   *ProducerOption
	writer *kafka.Writer
}

func NewProducer(opt *ProducerOption) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: strings.Split(opt.Brokers, ","),
	})
	return &Producer{opts: opt, writer: writer}, nil
}

type Message = kafka.Message

func (th *Producer) WriteMessage(ctx context.Context, msgs ...Message) error {
	return th.writer.WriteMessages(ctx, msgs...)
}
