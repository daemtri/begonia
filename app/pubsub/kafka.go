package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/daemtri/begonia/driver/kafka"
)

type kafkaMessage struct {
	msg *kafka.Message
}

func (m *kafkaMessage) Topic() string {
	return m.msg.Topic
}

func (m *kafkaMessage) Value() []byte {
	return m.msg.Value
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

func (ks *kafkaPublisher) Publish(ctx context.Context, msg Message) error {
	return ks.producer.WriteMessage(ctx, kafka.Message{
		Topic: msg.Topic(),
		Key:   []byte(fmt.Sprintf("%d", time.Now().UnixNano())),
		Value: msg.Value(),
	})
}

type kafkaMessageReader struct {
	prev   *kafkaMessage
	reader *kafka.Reader
}

func (km *kafkaMessageReader) Next() (Message, error) {
	if km.prev != nil {
		km.reader.CommitMessages(context.Background(), *km.prev.msg)
	}
	msg, err := km.reader.FetchMessage(context.Background())
	if err != nil {
		return nil, err
	}
	km.prev = &kafkaMessage{
		msg: &msg,
	}

	return km.prev, nil
}

type kafkaSubscriber struct {
	consumer *kafka.Consumer
}

func NewKafkaSubscriber(consumer *kafka.Consumer) Subscriber {
	return &kafkaSubscriber{
		consumer: consumer,
	}
}

func (ks *kafkaSubscriber) Subscribe(ctx context.Context, topic ...string) MessageReader {
	rd := ks.consumer.NewReader(ctx, topic...)
	return &kafkaMessageReader{
		reader: rd,
	}
}
