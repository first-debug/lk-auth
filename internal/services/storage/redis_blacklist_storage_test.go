package storage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"auth-service/internal/domain/models"
	"auth-service/internal/services/jwt"
	storagepkg "auth-service/internal/services/storage"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var user = models.User{
	Email:   "test@mail.com",
	Version: 1,
	Role:    "user",
}

func getJWTService() (jwt.JWTService, error) {
	return jwt.NewJWTServiceImpl(
		[]byte("a-string-secret-at-least-256-bits-long"),
		time.Duration(time.Minute*15),
		time.Duration(time.Hour*24),
	)
}

func getRedisBlackListStorage(jwtService jwt.JWTService) *storagepkg.RedisBlackListStorage {
	ctx := context.Background()
	return storagepkg.NewRedisBlackListStorage(ctx,
		&redis.Options{
			Addr:     "192.168.0.175:6379",
			Password: "",
			DB:       0,
			Protocol: 2,
		},
		jwtService,
	)
}

func TestRedisBlackListStorage_AddToken(t *testing.T) {
	jwtService, _ := getJWTService()
	storage := getRedisBlackListStorage(jwtService)
	token, _ := jwtService.CreateAccessToken(user)

	err := storage.AddTokens(token)
	assert.Nil(t, err, "Add token is failed")
}

func TestRedisBlackListStorage_IsAllowed(t *testing.T) {
	jwtSrevice, _ := getJWTService()
	storage := getRedisBlackListStorage(jwtSrevice)
	token, _ := jwtSrevice.CreateAccessToken(user)

	storage.AddTokens(token)

	t.Run("IsAllowed true", func(t *testing.T) {
		allowed, err := storage.IsAllowed(token)
		fmt.Println(allowed)

		assert.Nil(t, err, "IsAllowed failed: ", err.Error())

		assert.Equal(t, true, allowed, "expected token to be allowed (not in blacklist)")
	})

	t.Run("IsAllowed false", func(t *testing.T) {
		allowed, err := storage.IsAllowed("not-exist-token")
		if err != nil {
			t.Errorf("IsAllowed failed: %v", err)
		}
		if allowed {
			t.Error("expected token to not be allowed (not in blacklist)")
		}
	})
}
