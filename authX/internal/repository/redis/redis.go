package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisDB struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisDB(ctx context.Context, addr, password string, db int, logger *zap.Logger) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	logger.Info("Redis connected", zap.String("addr", addr))

	return &RedisDB{
		client: client,
		logger: logger,
	}, nil
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}