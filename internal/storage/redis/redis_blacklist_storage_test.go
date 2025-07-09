package redis_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"lk-auth/internal/domain/model"
	"lk-auth/internal/service/jwt"
	"lk-auth/internal/storage"
	redispkg "lk-auth/internal/storage/redis"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var user = model.User{
	Email:   "test@mail.com",
	Version: 1,
	Role:    "user",
}

func getJWTService() (jwt.JWTService, error) {
	return jwt.NewJWTServiceImpl(
		[]byte("a-string-secret-at-least-256-bits-long"),
		time.Duration(time.Minute*15),
		time.Duration(time.Hour*24),
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
	)
}

func getRedisBlackListStorage(jwtService jwt.JWTService) (storage.BlackListStorage, error) {
	ctx := context.Background()
	opt := &redis.Options{
		Addr:     "192.168.0.175:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	}

	cl := redis.NewClient(opt)
	cl.Del(ctx, "*")
	cl.Close()

	return redispkg.NewRedisBlackListStorage(
		ctx,
		&sync.WaitGroup{},
		opt,
		jwtService,
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
		time.Duration(time.Second*30),
	)
}

func TestRedisBlackListStorage_AddToken(t *testing.T) {
	jwtService, _ := getJWTService()
	storage, err := getRedisBlackListStorage(jwtService)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwtService.CreateAccessToken(user)

	time.Sleep(time.Second)
	err = storage.AddTokens(token)
	assert.Nil(t, err, "Add token is failed")
}

func TestRedisBlackListStorage_IsAllowed(t *testing.T) {
	jwtSrevice, _ := getJWTService()
	storage, err := getRedisBlackListStorage(jwtSrevice)
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwtSrevice.CreateAccessToken(user)

	t.Run("IsAllowed true", func(t *testing.T) {
		allowed, err := storage.IsAllowed(token)
		fmt.Println(allowed)

		assert.Nil(t, err, "IsAllowed failed")

		assert.Equal(t, true, allowed, "expected token to be allowed (not in blacklist)")
	})

	t.Run("IsAllowed false", func(t *testing.T) {
		storage.AddTokens(token)
		allowed, err := storage.IsAllowed(token)
		if err != nil {
			t.Errorf("IsAllowed failed: %v", err)
		}
		if allowed {
			t.Error("expected token to not be allowed (in blacklist)")
		}
	})
}
