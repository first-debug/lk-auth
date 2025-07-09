package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"lk-auth/internal/config"
	"lk-auth/internal/server"
	"lk-auth/internal/service/auth"
	"lk-auth/internal/service/jwt"

	"lk-auth/internal/storage"
	graphStorage "lk-auth/internal/storage/graph"
	redisStorage "lk-auth/internal/storage/redis"

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
	redisOpts, err := redis.ParseURL(cfg.Storages.Redis)
	if err != nil {
		return nil, err
	}
	jwtStorage, err := redisStorage.NewRedisJWTStorage(
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

	blackListStorage, err := redisStorage.NewRedisBlackListStorage(
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

	userStorage := graphStorage.NewGraphQLUserStorage(
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
