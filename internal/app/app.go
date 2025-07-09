package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/first-debug/lk-auth/internal/config"
	"github.com/first-debug/lk-auth/internal/domain/models"
	"github.com/first-debug/lk-auth/internal/libs/hash"
	sl "github.com/first-debug/lk-auth/internal/libs/logger"
	"github.com/first-debug/lk-auth/internal/server"
	"github.com/first-debug/lk-auth/internal/services/auth"
	"github.com/first-debug/lk-auth/internal/services/jwt"
	"github.com/first-debug/lk-auth/internal/services/storage"

	"github.com/redis/go-redis/v9"
)

type App struct {
	log    *slog.Logger
	server *server.Server
	cfg    *config.Config

	jwtStorage       storage.JWTStorage
	blacklistStorage storage.BlackListStorage
	userStorage      storage.UserStorage
}

func New(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, log *slog.Logger, isShuttingDown *atomic.Bool) (*App, error) {
	// Инициализация зависимостей

	// JWT сервис
	jwtService, err := jwt.NewJWTServiceImpl(
		[]byte(cfg.SecretPhrase),
		cfg.TTL.Access,
		cfg.TTL.Refresh,
		log,
	)
	if err != nil {
		return nil, err
	}

	// Хранилища
	// RedisJWTStorage
	redisOpts, err := redis.ParseURL(cfg.Storages.Redis)
	if err != nil {
		return nil, err
	}
	jwtStorage, err := storage.NewRedisJWTStorage(
		ctx,
		wg,
		redisOpts,
		cfg.TTL.Refresh,
		log,
		cfg.PingTime,
	)
	if err != nil {
		return nil, err
	}

	// RedisBlackListStorage
	blackListStorage, err := storage.NewRedisBlackListStorage(
		ctx,
		wg,
		redisOpts,
		jwtService,
		log,
		cfg.PingTime,
	)
	if err != nil {
		return nil, err
	}

	// RedisUserStorage
	userStorage, err := storage.NewRedisUserStorage(
		ctx,
		wg,
		redisOpts,
		log,
		cfg.PingTime,
	)

	authService := auth.NewAuthServiceImpl(
		jwtService,
		blackListStorage,
		jwtStorage,
		userStorage,
		log,
	)

	hashPwd, err := hash.HashPassword("password")
	if err != nil {
		log.Error("cannot get hash from string", sl.Err(err))
	} else {
		userStorage.AddUsers(models.User{
			Email:        "e@e.com",
			PasswordHash: string(hashPwd),
			Role:         "student",
			Version:      1,
		})
	}

	srv := server.NewServer(ctx, authService, log, isShuttingDown)

	return &App{
		log:              log,
		server:           srv,
		cfg:              cfg,
		jwtStorage:       jwtStorage,
		blacklistStorage: blackListStorage,
		userStorage:      userStorage,
	}, nil
}

func (a *App) Run() error {
	a.log.Info("Запуск HTTP сервера по адресу '" + a.cfg.URL + ":" + a.cfg.Port + "'...")
	return a.server.Start(a.cfg.Env, a.cfg.URL+":"+a.cfg.Port)
}

func (a *App) ShutDown(shutDownCtx context.Context) error {
	if a == nil {
		return errors.New("App instance is nil")
	}

	err := errors.Join(
		a.server.ShutDown(shutDownCtx),
		a.jwtStorage.ShutDown(shutDownCtx),
		a.blacklistStorage.ShutDown(shutDownCtx),
		a.userStorage.ShutDown(shutDownCtx),
	)
	return err
}
