package app

import (
	"auth-service/internal/config"
	"auth-service/internal/domain/models"
	"auth-service/internal/libs/hash"
	sl "auth-service/internal/libs/logger"
	"auth-service/internal/server"
	"auth-service/internal/services/auth"
	"auth-service/internal/services/jwt"
	"auth-service/internal/services/storage"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type App struct {
	log    *slog.Logger
	server *server.Server
	cfg    *config.Config
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
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
	redisOpts, err := redis.ParseURL(cfg.Storages.RedisJWT)
	if err != nil {
		return nil, err
	}
	jwtStorage, err := storage.NewRedisJWTStorage(
		ctx,
		redisOpts,
		cfg.TTL.Refresh,
		log,
	)
	if err != nil {
		return nil, err
	}

	// RedisBlackListStorage
	redisOpts, err = redis.ParseURL(cfg.Storages.RedisBlackList)
	if err != nil {
		return nil, err
	}
	blackListStorage, err := storage.NewRedisBlackListStorage(ctx,
		redisOpts,
		jwtService,
		log,
	)
	if err != nil {
		return nil, err
	}

	// RedisUserStorage
	redisOpts, err = redis.ParseURL(cfg.Storages.RedisUser)
	if err != nil {
		return nil, err
	}
	userStorage, err := storage.NewRedisUserStorage(ctx,
		redisOpts,
		log,
	)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if cfg.Env != "prod" {
		passHash, err := hash.HashPassword("password")
		if err != nil {
			log.Debug("Inicialize mock DB", sl.Err(err))
		}
		userStorage.AddUsers(models.User{
			Email:        "example@example.com",
			PasswordHash: string(passHash), // password: password
			Role:         "student",
			Version:      1,
		})
	}

	authService := auth.NewAuthServiceImpl(
		jwtService,
		blackListStorage,
		jwtStorage,
		userStorage,
		log,
	)

	srv := server.NewServer(authService, log)

	return &App{
		log:    log,
		server: srv,
		cfg:    cfg,
	}, nil
}

func (a *App) Run() error {
	a.log.Info("Запуск HTTP сервера по адресу '" + a.cfg.URL + ":" + a.cfg.Port + "'...")
	// TODO: реализовать graceful shutdown
	a.server.Start(a.cfg.Env, a.cfg.URL+":"+a.cfg.Port)
	return nil
}
