//go:build integration

package redis_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"lk-auth/internal/domain/model"
	jwtpkg "lk-auth/internal/service/jwt"
	"lk-auth/internal/storage"
	redispkg "lk-auth/internal/storage/redis"
	mock "lk-auth/internal/testutil/mock/jwt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var user = model.User{
	Email:   "test@mail.com",
	Version: 1,
	Role:    "user",
}

func getRedisBlackListStorage(jwtService jwtpkg.JWTService) (storage.BlackListStorage, error) {
	ctx := context.Background()

	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}
	if opt.DB == 0 {
		return nil, errors.New("test enviroment! don't use 0 db")
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
	jwtService := &mock.MockJWTService{}
	storage, err := getRedisBlackListStorage(jwtService)

	if err != nil {
		t.Fatal(err)
	}

	access := "access"
	jwtService.On("CreateAccessToken", user).Return(access, nil).Once()
	jwtService.On("GetTokenClaims", access).Return(
		jwt.MapClaims{
			"exp": float64(time.Now().Add(time.Minute).Unix()),
		},
		nil,
	).Once()

	_, err = jwtService.CreateAccessToken(user)

	assert.Nil(t, err)

	err = storage.AddTokens(access)
	assert.Nil(t, err, "Adding token is failed")

	jwtService.AssertExpectations(t)
}

func TestRedisBlackListStorage_IsAllowed(t *testing.T) {
	jwtService := &mock.MockJWTService{}
	storage, err := getRedisBlackListStorage(jwtService)

	if err != nil {
		t.Fatal(err)
	}

	t.Run("IsAllowed true", func(t *testing.T) {
		allowed, err := storage.IsAllowed("allowed")

		assert.Nil(t, err, "IsAllowed failed")

		assert.Equal(t, true, allowed, "expected token to be allowed (not in blacklist)")
	})

	t.Run("IsAllowed false", func(t *testing.T) {
		disallowed := "disallowed"
		jwtService.On("GetTokenClaims", disallowed).Return(
			jwt.MapClaims{
				"exp": float64(time.Now().Add(time.Minute).Unix()),
			},
			nil,
		).Once()
		storage.AddTokens(disallowed)
		allowed, err := storage.IsAllowed(disallowed)
		if err != nil {
			t.Errorf("IsAllowed failed: %v", err)
		}
		if allowed {
			t.Error("expected token to not be allowed (in blacklist)")
		}
	})
}
