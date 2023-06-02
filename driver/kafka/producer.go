package kafka

import (
	"fmt"
	"strings"

	"git.bianfeng.com/stars/wegame/wan/wanx/app"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"github.com/Shopify/sarama"
)

// Producer 消息生产组件
type Producer struct {
	Option *ProducerOption
	config *sarama.Config

	topicAdapterFunc TopicAdapterFunc
	asyncProducer    sarama.AsyncProducer
}

type ProducerOption struct {
	Brokers string `flag:"brokers" default:"127.0.0.1:9092" usage:"kafka brokers"`
}

func NewAsyncProducer(_ box.Context, opt *ProducerOption) (*Producer, error) {
	config := sarama.NewConfig()

	producer, err := sarama.NewAsyncProducer(strings.Split(opt.Brokers, ","), config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		asyncProducer:    producer,
		Option:           opt,
		config:           config,
		topicAdapterFunc: topicWithNamespace,
	}, nil
}

type TopicAdapterFunc func(topic string) string

func topicWithNamespace(topic string) string {
	return namespaceTopic(app.GetNamespace(), topic)
}

func TopicOnlyAdapterFunc(topic string) string {
	return topic
}

func (th *Producer) SetTopicAdapterFunc(fn TopicAdapterFunc) {
	th.topicAdapterFunc = fn
}

func (th *Producer) Send(producerMessage *sarama.ProducerMessage) {
	producerMessage.Topic = th.topicAdapterFunc(producerMessage.Topic)
	th.asyncProducer.Input() <- producerMessage
}

func (th *Producer) SendMessage(topic string, message []byte) {
	th.Send(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	})
}

func (th *Producer) SendMessageWithoutNamespace(topic string, message []byte) {
	th.SendWithoutNamespace(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	})
}

func (th *Producer) SendToPartition(topic string, partition int32, message []byte) {
	th.Send(&sarama.ProducerMessage{
		Topic:     topic,
		Partition: partition,
		Value:     sarama.ByteEncoder(message),
	})
}

func (th *Producer) SendString(topic string, key string, message string) {
	th.Send(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	})
}

func (th *Producer) SendWithoutNamespace(producerMessage *sarama.ProducerMessage) {
	th.asyncProducer.Input() <- producerMessage
}

func (th *Producer) SendMessageWithNamespace(namespace, topic string, message []byte) {
	th.SendWithoutNamespace(&sarama.ProducerMessage{
		Topic: namespaceTopic(namespace, topic),
		Value: sarama.ByteEncoder(message),
	})
}

func (th *Producer) SendToPartitionWithNamespace(namespace, topic string, partition int32, message []byte) {
	th.SendWithoutNamespace(&sarama.ProducerMessage{
		Topic:     namespaceTopic(namespace, topic),
		Partition: partition,
		Value:     sarama.ByteEncoder(message),
	})
}

func (th *Producer) SendStringWithNamespace(namespace, topic, key string, message string) {
	th.SendWithoutNamespace(&sarama.ProducerMessage{
		Topic: namespaceTopic(namespace, topic),
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	})
}

func namespaceTopic(namespace, topic string) string {
	return fmt.Sprintf("%s.%s", namespace, topic)
}
