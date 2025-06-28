package storage

import (
	"auth-service/internal/services/jwt"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBlackListStorage struct {
	ctx        context.Context
	client     *redis.Client
	jwtService jwt.JWTService
}

func NewRedisBlackListStorage(ctx context.Context, options *redis.Options, jwtService jwt.JWTService) *RedisBlackListStorage {
	return &RedisBlackListStorage{
		ctx:        ctx,
		client:     redis.NewClient(options),
		jwtService: jwtService,
	}
}

func (s *RedisBlackListStorage) AddTokens(tokens ...string) error {
	for _, token := range tokens {
		claims, err := s.jwtService.GetTokenClaims(token)
		if err != nil {
			return err
		}
		expClaim, exists := claims["exp"]
		if !exists {
			return errors.New("token missing expiration claim")
		}
		exp, ok := expClaim.(float64)
		if !ok {
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

	return response == 1, nil
}
