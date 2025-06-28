package storage_test

import (
	storagepkg "auth-service/internal/services/storage"
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func getRedisJWTStorage() *storagepkg.RedisJWTStorage {
	ctx := context.Background()
	return storagepkg.NewRedisJWTStorage(ctx, &redis.Options{
		Addr:     "192.168.0.175:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	},
		time.Duration(time.Minute*15),
	)
}

func TestRedisJWTStorage_AddPair(t *testing.T) {
	client := getRedisJWTStorage()
	st1, st2 := "access", "refresh"
	err := client.AddPair(st1, st2)
	assert.Nil(t, err)
}

func TestRedisJWTStorage_GetAccessByRefresh(t *testing.T) {
	client := getRedisJWTStorage()
	st1, st2 := "access", "refresh"
	client.AddPair(st1, st2)

	res, err := client.GetAccessByRefresh(st2)

	assert.Nil(t, err)
	assert.Equal(t, st1, res)

	res, err = client.GetAccessByRefresh(st2)
	assert.NotNil(t, err)
	assert.Equal(t, res, "")
}
