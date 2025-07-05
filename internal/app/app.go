package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/first-debug/lk-auth/internal/config"
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
	redisOpts, err := redis.ParseURL(cfg.Storages.JWT)
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
	redisOpts, err = redis.ParseURL(cfg.Storages.BlackList)
	if err != nil {
		return nil, err
	}
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

	// GraphQLUserStorage
	userStorage := storage.NewGraphQLUserStorage(
		ctx,
		wg,
		log,
		cfg.PingTime,
		cfg.Storages.Users,
	)

	authService := auth.NewAuthServiceImpl(
		jwtService,
		blackListStorage,
		jwtStorage,
		userStorage,
		log,
	)

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
	a.server.Start(a.cfg.Env, a.cfg.URL+":"+a.cfg.Port)
	return nil
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
