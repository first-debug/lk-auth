// Точка входа
package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"
	sl "auth-service/internal/libs/logger"
	"context"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg)

	log.Info("Start application...")

	ctx := context.Background()

	// Инициализация и запуск приложения
	a, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("Не удалось инициализировать приложение", sl.Err(err))
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		log.Error("Ошибка при запуске приложения", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger

	// If logger.level varable is not set set [slog.Level] to DEBUG for "local" and "dev" and INFO for "prod"
	if cfg.Logger.Level == nil {
		var level slog.Level
		if cfg.Env != "prod" {
			level = slog.LevelDebug.Level()
		} else {
			level = slog.LevelInfo.Level()
		}
		cfg.Logger.Level = &level
	}

	switch cfg.Env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	}

	return log
}
