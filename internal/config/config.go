// Парсер файла конфигурации
package config

import (
	"flag"
	"log/slog"
	"os"
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

	URL  string `yaml:"url" env-default:""`
	Port string `yaml:"port" env-default:"80"`
	TTL  struct {
		Access  time.Duration `yaml:"access" env-default:"15m"`
		Refresh time.Duration `yaml:"refresh" env-default:"1h"`
	} `yaml:"ttl"`
	Logger struct {
		Level        *slog.Level `yaml:"level"`
		ShowPathCall bool        `yaml:"show_path_call" env-default:"false"`
	} `yaml:"logger"`
	PingTime time.Duration `yaml:"ping_time" env-default:"1m"`
	Shutdown struct {
		Period     time.Duration `yaml:"period" env-default:"15s"`
		HardPeriod time.Duration `yaml:"hard_period" env-default:"3s"`
	} `yaml:"shutdown"`
	Readiness struct {
		DrainDelay time.Duration `yaml:"drain_delay" env-default:"5s"`
	} `yaml:"readiness"`
}

// По соглашению, функции с префиксом Must вместо возвращения ошибок создают панику.
// Используйте их с осторожностью.
func MustLoad() *Config {
	godotenv.Load()

	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic("config file does not exist: " + configPath)
	}

	cfg := &Config{}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		panic(err.Error())
	}

	return cfg
}

// fetchConfigPath извлекает путь до файла конфигурации из аргументов командной строки или переменнх окружения
// приоритет flag > env > default
// Дефолтное значение - пустая строка
func fetchConfigPath() (res string) {
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	if res == "" {
		res = "config/config_local.yml"
	}
	return
}
