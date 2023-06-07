package app

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

// GetDB  获取数据库
func GetDB(ctx context.Context, name string) *sql.DB {
	mysql.NewConfig()
	// mr := moduleRuntimeFromCtx(ctx)
	return global.dbClient.MustGetOrInit(name, func() *sql.DB {
		return nil
	})
}

// GetRedis 获取redis
func GetRedis(ctx context.Context, name string) *redis.Client {
	// mr := moduleRuntimeFromCtx(ctx)
	return global.redisClient.MustGetOrInit(name, func() *redis.Client {
		return nil
	})
}
