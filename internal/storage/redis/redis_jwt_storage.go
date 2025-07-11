package redis

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	sl "lk-auth/internal/libs/logger"
	"lk-auth/internal/storage"

	"github.com/redis/go-redis/v9"
)

const jwtPref = "auth:jwt:"

type RedisJWTStorage struct {
	ctx    context.Context
	client *redis.Client
	ttl    time.Duration
	log    *slog.Logger
}

func NewRedisJWTStorage(ctx context.Context, wg *sync.WaitGroup, options *redis.Options, ttl time.Duration, log *slog.Logger, pingTime time.Duration) (storage.JWTStorage, error) {
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdin, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
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
					log.Error("RedisJWTStorage didn't answer", "url", options.Addr)
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
	err := s.client.Set(s.ctx, jwtPref+refresh, access, s.ttl).Err()
	if err != nil {
		s.log.Error("Cannot add pair", sl.Err(err))
	}

	return err
}

func (s *RedisJWTStorage) GetAccessByRefresh(refresh string) (string, error) {
	res, err := s.client.Get(s.ctx, jwtPref+refresh).Result()
	if err != nil {
		return "", err
	}
	if err = s.client.Del(s.ctx, jwtPref+refresh).Err(); err != nil {
		s.log.Error("Cannot delete pair", sl.Err(err))
		return "", err
	}

	return res, nil
}

func (s *RedisJWTStorage) ShutDown(shutDownCtx context.Context) error {
	return s.client.Close()
}
