package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisJWTStorage struct {
	ctx    context.Context
	client *redis.Client
	ttl    time.Duration
}

func NewRedisJWTStorage(ctx context.Context, option *redis.Options, ttl time.Duration) *RedisJWTStorage {
	return &RedisJWTStorage{
		ttl:    ttl,
		client: redis.NewClient(option),
		ctx:    ctx,
	}
}

func (s *RedisJWTStorage) AddPair(access string, refresh string) error {
	return s.client.Set(s.ctx, refresh, access, s.ttl).Err()
}

func (s *RedisJWTStorage) GetAccessByRefresh(refresh string) (string, error) {
	res, err := s.client.Get(s.ctx, refresh).Result()
	if err != nil {
		return "", err
	}
	if err = s.client.Del(s.ctx, refresh).Err(); err != nil {
		return "", err
	}

	return res, nil
}
