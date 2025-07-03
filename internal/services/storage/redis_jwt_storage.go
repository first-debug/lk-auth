package storage

import (
	sl "auth-service/internal/libs/logger"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisJWTStorage struct {
	ctx    context.Context
	client *redis.Client
	ttl    time.Duration
	log    *slog.Logger
}

func NewRedisJWTStorage(ctx context.Context, wg *sync.WaitGroup, options *redis.Options, ttl time.Duration, log *slog.Logger, pingTime time.Duration) (JWTStorage, error) {
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(pingTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Debug("RedisJWTStorage ping goroutine stopped")
				wg.Done()
				return
			case <-ticker.C:
				if err := client.Ping(ctx).Err(); err != nil {
					log.Error("RedisUserStorage didn't answer", "url", options.Addr)
				}
			}
		}
	}()

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

func (s *RedisJWTStorage) ShutDown(shutDownCtx context.Context) error {
	return s.client.Close()
}
