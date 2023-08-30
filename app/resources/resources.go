package resources

import (
	"context"
	"fmt"

	"github.com/daemtri/begonia/app/pubsub"
	"github.com/daemtri/begonia/driver/db"
	"github.com/daemtri/begonia/driver/kafka"
	"github.com/daemtri/begonia/driver/redis"
	"github.com/daemtri/begonia/pkg/helper"
	"github.com/daemtri/begonia/runtime/component"
)

type DBConfig struct {
	Name           string `json:"name"`
	Driver         string `json:"driver"`
	DataSourceName string `json:"dsn"`
}

type RedisConfig struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type PubSubConfig struct {
	Name    string `json:"name"`
	Driver  string `json:"driver"`
	Brokers string `json:"brokers"`
	Group   string `json:"group"` // 仅在消费者中有效
}

type Config struct {
	Redis  []RedisConfig  `yaml:"redis"`
	DB     []DBConfig     `yaml:"db"`
	PubSub []PubSubConfig `yaml:"pubsub"`
}

func (r *Config) GetDBConfig(name string) *DBConfig {
	for _, cfg := range r.DB {
		if cfg.Name == name {
			return &cfg
		}
	}
	return nil
}

func (r *Config) GetRedisConfig(name string) *RedisConfig {
	for _, cfg := range r.Redis {
		if cfg.Name == name {
			return &cfg
		}
	}
	return nil
}

func (r *Config) GetPubSubConfig(name string) *PubSubConfig {
	for _, cfg := range r.PubSub {
		if cfg.Name == name {
			return &cfg
		}
	}
	return nil
}

type Manager struct {
	configor component.Configurator
	config   *Config

	dbClients    helper.OnceMap[string, *db.Database]
	redisClients helper.OnceMap[string, *redis.Redis]
	publisher    helper.OnceMap[string, pubsub.Publisher]
	subscriber   helper.OnceMap[string, pubsub.Subscriber]
}

func NewManager(ctx context.Context, configor component.Configurator) (*Manager, error) {
	m := &Manager{configor: configor}
	return m, m.init(ctx)
}

func (m *Manager) init(ctx context.Context) error {
	cfg, err := m.configor.ReadConfig(ctx, "resources")
	if err != nil {
		return nil
	}
	var config Config
	if err := cfg.Decode(&config); err != nil {
		return err
	}
	m.config = &config
	return nil
}

func (m *Manager) GetDB(ctx context.Context, name string) (*db.Database, error) {
	return m.dbClients.GetOrInit(name, func() (*db.Database, error) {
		cfg := m.config.GetDBConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("db name %s config not found", name)
		}
		db, err := db.NewDB(&db.Options{
			DriverName: cfg.Driver,
			DSN:        cfg.DataSourceName,
		})
		if err != nil {
			return nil, fmt.Errorf("open db %s error: %w", name, err)
		}
		return db, nil
	})
}

func (m *Manager) GetRedis(ctx context.Context, name string) (*redis.Redis, error) {
	return m.redisClients.GetOrInit(name, func() (*redis.Redis, error) {
		cfg := m.config.GetRedisConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("redis name %s config not found", name)
		}
		return redis.NewRedis(ctx, &redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	})
}

func (m *Manager) GetMsgSubscriber(ctx context.Context, name string) (pubsub.Subscriber, error) {
	return m.subscriber.GetOrInit(name, func() (pubsub.Subscriber, error) {
		cfg := m.config.GetPubSubConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("kafka name %s config not found", name)
		}
		c, err := kafka.NewConsumer(&kafka.ConsumerOption{
			Brokers: cfg.Brokers,
			Group:   cfg.Group,
		})
		if err != nil {
			return nil, err
		}
		return pubsub.NewKafkaSubscriber(c), nil
	})
}

func (m *Manager) GetMsgPublisher(ctx context.Context, name string) (pubsub.Publisher, error) {
	return m.publisher.GetOrInit(name, func() (pubsub.Publisher, error) {
		cfg := m.config.GetPubSubConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("kafka name %s config not found", name)
		}
		p, err := kafka.NewProducer(&kafka.ProducerOption{
			Brokers: cfg.Brokers,
		})
		if err != nil {
			return nil, err
		}
		return pubsub.NewKafkaPublisher(p), nil
	})
}
