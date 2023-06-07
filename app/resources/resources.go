package resources

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/helper"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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
}

type KafkaConfig struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type Config struct {
	Redis []RedisConfig `yaml:"redis"`
	DB    []DBConfig    `yaml:"db"`
	Kafka []KafkaConfig `yaml:"kafka"`
}

func (r *Config) GetDBConfig(name string) *DBConfig {
	for _, db := range r.DB {
		if db.Name == name {
			return &db
		}
	}
	return nil
}

func (r *Config) GetRedisConfig(name string) *RedisConfig {
	for _, redis := range r.Redis {
		if redis.Name == name {
			return &redis
		}
	}
	return nil
}

func (r *Config) GetKafkaConfig(name string) *KafkaConfig {
	for _, kafka := range r.Kafka {
		if kafka.Name == name {
			return &kafka
		}
	}
	return nil
}

type Manager struct {
	configor component.Configuration
	config   *Config
	once     sync.Once

	dbClients    helper.OnceMap[string, *sql.DB]
	redisClients helper.OnceMap[string, *redis.Client]
	kafkaClients helper.OnceMap[string, *kafka.Conn]
}

func NewManager(configor component.Configuration) (*Manager, error) {
	return &Manager{configor: configor}, nil
}

func (r *Manager) init(ctx context.Context) {
	r.once.Do(func() {})
}

func (r *Manager) GetDB(ctx context.Context, name string) (*sql.DB, error) {
	r.init(ctx)
	return r.dbClients.GetOrInit(name, func() (*sql.DB, error) {
		cfg := r.config.GetDBConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("db name %s config not found", name)
		}
		db, err := sql.Open(cfg.Driver, cfg.DataSourceName)
		if err != nil {
			return nil, fmt.Errorf("open db %s error: %w", name, err)
		}
		return db, nil
	})
}

func (r *Manager) GetRedis(ctx context.Context, name string) (*redis.Client, error) {
	r.init(ctx)
	return r.redisClients.GetOrInit(name, func() (*redis.Client, error) {
		cfg := r.config.GetRedisConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("redis name %s config not found", name)
		}
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
		})
		return client, nil
	})
}

func (r *Manager) GetKafka(ctx context.Context, name string) (*kafka.Conn, error) {
	r.init(ctx)
	return r.kafkaClients.GetOrInit(name, func() (*kafka.Conn, error) {
		cfg := r.config.GetKafkaConfig(name)
		if cfg == nil {
			return nil, fmt.Errorf("kafka name %s config not found", name)
		}
		conn, err := kafka.Dial("tcp", cfg.Addr)
		if err != nil {
			return nil, fmt.Errorf("dial kafka %s error: %w", name, err)
		}
		return conn, nil
	})
}
