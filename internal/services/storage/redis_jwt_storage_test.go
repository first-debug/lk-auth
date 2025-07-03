package storage_test

import (
	storagepkg "auth-service/internal/services/storage"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func getRedisJWTStorage() (storagepkg.JWTStorage, error) {
	ctx := context.Background()
	return storagepkg.NewRedisJWTStorage(ctx, &redis.Options{
		Addr:     "192.168.0.175:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	},
		time.Duration(time.Minute*15),
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		),
		time.Duration(time.Second*30),
	)
}

func TestRedisJWTStorage_AddPair(t *testing.T) {
	client, err := getRedisJWTStorage()
	if err != nil {
		t.Fatal(err)
	}
	st1, st2 := "access", "refresh"
	err = client.AddPair(st1, st2)
	assert.Nil(t, err)
}

func TestRedisJWTStorage_GetAccessByRefresh(t *testing.T) {
	client, err := getRedisJWTStorage()
	if err != nil {
		t.Fatal(err)
	}
	st1, st2 := "access", "refresh"
	client.AddPair(st1, st2)

	res, err := client.GetAccessByRefresh(st2)

	assert.Nil(t, err)
	assert.Equal(t, st1, res)

	res, err = client.GetAccessByRefresh(st2)
	assert.NotNil(t, err)
	assert.Equal(t, res, "")
}
