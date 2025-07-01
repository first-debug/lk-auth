package storage

import (
	sl "auth-service/internal/libs/logger"
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisJWTStorage struct {
	ctx    context.Context
	client *redis.Client
	ttl    time.Duration
	log    *slog.Logger
}

func NewRedisJWTStorage(ctx context.Context, option *redis.Options, ttl time.Duration, log *slog.Logger) (JWTStorage, error) {
	client := redis.NewClient(option)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisJWTStorage{
		ttl:    ttl,
		client: client,
		ctx:    ctx,
		log:    log,
	}, nil
}

func (s *RedisJWTStorage) AddPair(access string, refresh string) error {
	err := s.client.Set(s.ctx, refresh, access, s.ttl).Err()
	if err != nil {
		s.log.Error("Cannot add pair", sl.Err(err))
	}

	return err
}

func (s *RedisJWTStorage) GetAccessByRefresh(refresh string) (string, error) {
	// TODO: изучить что и в каком случае возвращает метод Result() и
	// в зависимости от этого улучшить код
	res, err := s.client.Get(s.ctx, refresh).Result()
	if err != nil {
		return "", err
	}
	if err = s.client.Del(s.ctx, refresh).Err(); err != nil {
		s.log.Error("Cannot delete pair", sl.Err(err))
		return "", err
	}

	return res, nil
}
