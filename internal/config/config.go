// Парсер файла конфигурации
package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
}

// По соглашению, функции с префиксом Must вместо возвращения ошибок создают панику.
// Используйте их с осторожностью.
func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err.Error())
	}

	return &cfg
}

// fetchConfigPath извлекает путь до файла конфигурации из аргументов командной строки или переменнх окружения
// приоритет falg > env > default
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
