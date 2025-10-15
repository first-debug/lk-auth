package redis

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"lk-auth/internal/service/jwt"
	"lk-auth/internal/storage"

	"github.com/redis/go-redis/v9"
)

const blacklistPref = "auth:blacklist:"

type RedisBlackListStorage struct {
	ctx        context.Context
	client     *redis.Client
	jwtService jwt.JWTService

	log *slog.Logger
}

func NewRedisBlackListStorage(ctx context.Context, wg *sync.WaitGroup, options *redis.Options, jwtService jwt.JWTService, log *slog.Logger, pingTime time.Duration) (storage.BlackListStorage, error) {
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
				log.Debug("RedisBlackListStorage ping goroutine stopped")
				wg.Done()
				return
			case <-ticker.C:
				if err := client.Ping(ctx).Err(); err != nil {
					log.Error("RedisBlackListStorage didn't answer", "url", options.Addr)
				}
			}
		}
	}()

	return &RedisBlackListStorage{
		ctx:        ctx,
		client:     client,
		jwtService: jwtService,
		log:        log,
	}, nil
}

func (s *RedisBlackListStorage) AddTokens(tokens ...string) error {
	for _, token := range tokens {
		claims, err := s.jwtService.GetTokenClaims(token)
		if err != nil {
			continue
		}
		expClaim := claims["exp"]

		exp, ok := expClaim.(float64)
		if !ok {
			s.log.Error("token expiration claim is not a number", "token", token)
			return errors.New("token expiration claim is not a number")
		}
		dur := time.Duration(int64(exp)-time.Now().Unix()) * time.Second
		err = s.client.Set(s.ctx, blacklistPref+token, true, dur).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisBlackListStorage) IsAllowed(token string) (bool, error) {
	response, err := s.client.Exists(s.ctx, blacklistPref+token).Result()
	if err != nil {
		return false, err
	}

	return response == 0, nil
}

func (s *RedisBlackListStorage) ShutDown(shutDownCtx context.Context) error {
	return s.client.Close()
}
