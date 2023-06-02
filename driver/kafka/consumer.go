package kafka

import (
	"context"
	"fmt"
	"strings"

	"git.bianfeng.com/stars/wegame/wan/wanx/app"
	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

var logger = logx.GetLogger("gf/driver/kafka")

// Consumer 消息消费组件
type Consumer struct {
	client sarama.Client
	config *sarama.Config
}

type ConsumerOption struct {
	Brokers string `flag:"brokers" default:"127.0.0.1:9092" usage:"Kafka bootstrap Brokers to connect to, as a comma separated list"`
	Version string `flag:"version" default:"2.1.1" usage:"Kafka cluster version"`
	Oldest  bool   `flag:"oldest" default:"false" usage:"Kafka consumer consume initial offset from oldest"`
}

func NewConsumer(_ box.Context, opt *ConsumerOption) (*Consumer, error) {
	config := sarama.NewConfig()
	// 使用option中的参数覆盖config
	if opt.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	v, err := sarama.ParseKafkaVersion(opt.Version)
	if err != nil {
		return nil, err
	}
	config.Version = v

	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	client, err := sarama.NewClient(strings.Split(opt.Brokers, ","), config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client: client,
		config: config,
	}, nil
}

// Consume
//
//	@Description: 以 group 为订单组，订阅 topics 中的消息
//	@receiver th
//	@param ctx
//	@param consumer 实现了 sarama.ConsumerGroupHandler 接口的消息处理器
func (th *Consumer) Consume(ctx context.Context, topics, group string, consumer sarama.ConsumerGroupHandler) (err error) {
	if len(topics) == 0 {
		err = errors.New("no topics given to be consumed, please set the -topics flag")
		return err
	}

	namespace := app.GetNamespace()

	tps := strings.Split(topics, ",")
	for i, tp := range tps {
		tps[i] = fmt.Sprintf("%s.%s", namespace, tp)
	}

	return th.ConsumeWithoutNamespace(ctx, strings.Join(tps, ","), group, consumer)
}

// ConsumeWithoutNamespace
//
//	@Description: 以 group 为订单组，订阅 topics 中的消息
//	@receiver th
//	@param ctx
//	@param consumer 实现了 sarama.ConsumerGroupHandler 接口的消息处理器
func (th *Consumer) ConsumeWithoutNamespace(
	ctx context.Context, topics, group string, consumer sarama.ConsumerGroupHandler) (err error) {
	group = fmt.Sprintf("%s.%s", app.GetNamespace(), group)
	logger.Info("kafka消息订阅", "group_name", group, "topics", topics)

	cg, err := sarama.NewConsumerGroupFromClient(group, th.client)
	if err != nil {
		return err
	}

	if len(topics) == 0 {
		err = errors.New("no topics given to be consumed, please set the -topics flag")
		return err
	}

	tps := strings.Split(topics, ",")
	go func() {
		for {
			err = cg.Consume(ctx, tps, consumer)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}
