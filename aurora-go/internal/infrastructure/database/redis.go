package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aurora-go/aurora/internal/config"
)

var RDB *redis.Client

// InitRedis 初始化 Redis 连接（对标 Java 版 RedisTemplate 全部数据结构操作）
func InitRedis(cfg *config.RedisConfig) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		panic("Failed to connect to Redis: " + err.Error())
	}

	slog.Info("Redis connected successfully",
		"addr", cfg.Addr(),
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RDB == nil {
		return nil
	}
	return RDB.Close()
}

// GetRedis 获取 Redis 客户端实例
func GetRedis() *redis.Client {
	return RDB
}
