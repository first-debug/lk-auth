package storage

import (
	"context"
	"errors"
	"lk-auth/internal/domain/models"
	"lk-auth/internal/libs/hash"
	sl "lk-auth/internal/libs/logger"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUserStorage struct {
	ctx    context.Context
	client *redis.Client
	log    *slog.Logger
}

func NewRedisUserStorage(ctx context.Context, wg *sync.WaitGroup, options *redis.Options, log *slog.Logger, pingTime time.Duration) (UserStorage, error) {
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
				log.Debug("RedisUserStorage ping goroutine stopped")
				wg.Done()
				return
			case <-ticker.C:
				if err := client.Ping(ctx).Err(); err != nil {
					log.Error("RedisUserStorage didn't answer", "url", options.Addr)
				}
			}
		}
	}()

	return &RedisUserStorage{
		client: client,
		ctx:    ctx,
		log:    log,
	}, nil
}

// from UserProvider interface
func (s *RedisUserStorage) Login(email, password string) (float64, string, error) {
	if email == "" || len(password) == 0 {
		s.log.Error("invalid input parameters")
		return -1, "", errors.New("invalid input parameters")
	}

	userInfo := models.User{}
	err := s.client.HGetAll(s.ctx, email).Scan(&userInfo)

	if err != nil {
		if err == redis.Nil {
			return -1, "", errors.New("user not found")
		}
		s.log.Error("database error", sl.Err(err))
		return -1, "", err
	}

	if len(userInfo.PasswordHash) == 0 || !hash.CheckPasswordHash([]byte(password), []byte(userInfo.PasswordHash)) {
		return -1, "", errors.New("incorrect email and password")
	}

	return userInfo.Version, userInfo.Role, nil
}

// from UserProvider interface
func (s *RedisUserStorage) IsVersionValid(email string, version float64) (bool, error) {
	if len(email) == 0 {
		return false, errors.New("email cannot be empty")
	}

	userInfo := models.User{}
	err := s.client.HGetAll(s.ctx, email).Scan(&userInfo)
	if err != nil {
		if err == redis.Nil {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return userInfo.Version == version, nil
}

// метод для добавления пользователей в базу данных
func (s *RedisUserStorage) AddUsers(users ...models.User) error {
	if len(users) == 0 {
		return nil
	}

	pipe := s.client.TxPipeline()
	for _, user := range users {
		pipe.HSet(s.ctx, user.Email, user)
	}
	_, err := pipe.Exec(s.ctx)
	if err != nil {
		s.log.Error("database error", sl.Err(err))
	}
	return err
}

func (s *RedisUserStorage) ShutDown(shutDownCtx context.Context) error {
	return s.client.Close()
}
