package storage

import (
	"auth-service/internal/services/jwt"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBlackListStorage struct {
	ctx        context.Context
	client     *redis.Client
	jwtService jwt.JWTService

	log *slog.Logger
}

func NewRedisBlackListStorage(ctx context.Context, options *redis.Options, jwtService jwt.JWTService, log *slog.Logger, pingTime time.Duration) (BlackListStorage, error) {
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(pingTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Debug("RedisBlackListStorage ping gorutine stopped")
				return
			case <-ticker.C:
				if err := client.Ping(ctx).Err(); err != nil {
					log.Error("RedisBlackListStorage didn't answer", "url", options.Addr)
				}
			}
		}
	}(ctx)

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
		err = s.client.Set(s.ctx, token, true, dur).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisBlackListStorage) IsAllowed(token string) (res bool, err error) {
	response, err := s.client.Exists(s.ctx, token).Result()
	if err != nil {
		return false, err
	}

	return response == 0, nil
}
