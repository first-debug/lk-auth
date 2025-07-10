package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"lk-auth/internal/app"
	"lk-auth/internal/config"
	sl "lk-auth/internal/libs/logger"

	"github.com/lmittmann/tint"
)

var isShuttingDown atomic.Bool

func main() {
	// The core principle of graceful shutdown is the same across all systems: Stop accepting new requests or messages, and give existing operations time to finish within a defined grace period.

	// Создаём основной контекст, который ждёт вызов сигналов на заверешние работы
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()

	log := setupLogger(cfg)

	log.Info("Start application...")

	// Создаём контекст runtime'а, в котором будут работать все компоненты приложения
	ongoingCtx, stopOngoingGracefull := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Инициализацируем приложение
	a, err := app.New(ongoingCtx, wg, cfg, log, &isShuttingDown)
	if err != nil {
		log.Error("Не удалось инициализировать приложение.", sl.Err(err))
		stop()
	}

	// Запускаем приложение
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-rootCtx.Done():
			return
		default:
			err = a.Run()
			if err != nil {
				log.Error("Ошибка при запуске приложения.", sl.Err(err))
				stop()
			}
		}
	}()

	// Ожидаем сигналы к завершению
	<-rootCtx.Done()
	// Устанавливаем флаг состояния isShuttingDown true, для оповещения внешних сервисов о завешении работы (см. [server.handleHealthz])
	isShuttingDown.Store(true)
	log.Info("Получен сигнал отключения, выключение...")

	// Ждём пока сообщение об изменении статуса приложения дойдёт до внешних сервисов
	if cfg.Env == "prod" {
		log.Info("Распростронение информации о завершении работы...")
		time.Sleep(cfg.Readiness.DrainDelay)
	}

	// Создаём контекст выключения с таймаутом, для ограничения его времени
	shutDownCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown.Period)
	defer cancel()

	// Вызываем метод для корректного завршения работы приложения
	// Метод ShutDown вернёт значение только если всё компоненты приложеня мягко завершат работу(например, http сервер - закроет все существующие подключения на момент выключения, RedisJWTStorage закроет - подключение к серверу Redis) или время выделенное на выключение истечёт
	err = a.ShutDown(shutDownCtx)
	stopOngoingGracefull()
	if err != nil {
		log.Error("Error shutting down", sl.Err(err))
		if cfg.Env == "prod" {
			time.Sleep(cfg.Shutdown.HardPeriod)
		}
	}
	wg.Wait()
	log.Info("Server shutdown gracefully.")
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
			tint.NewHandler(os.Stdout, &tint.Options{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	case "dev":
		log = slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	case "prod":
		log = slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				AddSource: cfg.Logger.ShowPathCall,
				Level:     cfg.Logger.Level,
			}),
		)
	}

	return log
}
