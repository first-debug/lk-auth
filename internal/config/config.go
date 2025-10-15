package config

import (
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	Storages struct {
		Redis string `env:"REDIS_URL" env-default:""`
		Users string `env:"USERS_URL" env-default:""`
	}
	SecretPhrase string `env:"SECRET_PHRASE" env-default:""`

	URL  string `env:"URL" env-default:""`
	Port string `env:"PORT" env-default:"80"`
	TTL  struct {
		Access  time.Duration `env:"TTL_ACCESS" env-default:"15m"`
		Refresh time.Duration `env:"TTL_REFRESH" env-default:"1h"`
	}
	Logger struct {
		Level        *slog.Level `env:"LOGGER_LEVEL" env-default:"INFO"`
		ShowPathCall bool        `env:"LOGGER_SHOW_PATH_CALL" env-default:"false"`
	} 
	PingTime time.Duration `env:"PING_TIME" env-default:"1m"`
	Shutdown struct {
		Period     time.Duration `env:"SHUTDOWN_PERIOD" env-default:"15s"`
		HardPeriod time.Duration `env:"SHUTDOWN_HARD_PERIOD" env-default:"3s"`
	} 
	Readiness struct {
		DrainDelay time.Duration `env:"READINESS_DRAIN_DELAY" env-default:"5s"`
	} 
}

// По соглашению, функции с префиксом Must вместо возвращения ошибок создают панику.
// Используйте их с осторожностью.
func MustLoad() *Config {
	cfg := &Config{}
	cfg.Logger.Level = new(slog.Level)

	godotenv.Load()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic(err.Error())
	}

	return cfg
}

