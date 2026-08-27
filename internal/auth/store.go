package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	refreshKeyPrefix   = "auth:refresh:"
	blacklistKeyPrefix = "auth:blacklist:"
)

type TokenStore interface {
	SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error
	GetRefresh(ctx context.Context, jti string) (string, error)
	RevokeRefresh(ctx context.Context, jti string) error
	BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessBlacklisted(ctx context.Context, jti string) (bool, error)
}

type RedisTokenStore struct {
	client *redis.Client
}

func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{client: client}
}

func (s *RedisTokenStore) SaveRefresh(ctx context.Context, jti, userID string, ttl time.Duration) error {
	return s.client.Set(ctx, refreshKeyPrefix+jti, userID, ttl).Err()
}

func (s *RedisTokenStore) GetRefresh(ctx context.Context, jti string) (string, error) {
	val, err := s.client.Get(ctx, refreshKeyPrefix+jti).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	return val, err
}

func (s *RedisTokenStore) RevokeRefresh(ctx context.Context, jti string) error {
	return s.client.Del(ctx, refreshKeyPrefix+jti).Err()
}

func (s *RedisTokenStore) BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return s.client.Set(ctx, blacklistKeyPrefix+jti, "1", ttl).Err()
}

func (s *RedisTokenStore) IsAccessBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := s.client.Exists(ctx, blacklistKeyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
