package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func (r *RedisDB) SaveUserToken(ctx context.Context, userID string, token string, ttl time.Duration) error {
	key := fmt.Sprintf("user_token:%s", userID)

	if err := r.client.Set(ctx, key, token, ttl).Err(); err != nil {
		r.logger.Error("Failed to save user token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("save user token: %w", err)
	}

	return nil
}

func (r *RedisDB) GetUserToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user_token:%s", userID)

	token, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		r.logger.Error("Failed to get user token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return "", fmt.Errorf("get user token: %w", err)
	}

	return token, nil
}

func (r *RedisDB) AddToBlacklist(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenID)

	if err := r.client.Set(ctx, key, "revoked", ttl).Err(); err != nil {
		r.logger.Error("Failed to add token to blacklist",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return fmt.Errorf("add to blacklist: %w", err)
	}

	return nil
}

func (r *RedisDB) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to check blacklist",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	return exists > 0, nil
}
