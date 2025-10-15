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

	"lk-auth/internal/storage"
	redispkg "lk-auth/internal/storage/redis"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func getRedisJWTStorage() (storage.JWTStorage, error) {
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

	return redispkg.NewRedisJWTStorage(
		ctx,
		&sync.WaitGroup{},
		opt,
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
	access, refresh := "access", "refresh"
	err = client.AddPair(access, refresh)
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
