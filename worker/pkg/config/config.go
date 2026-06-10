package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config espelha os nomes das variáveis de ambiente em maiúsculas
type Config struct {
	PORT                  string
	DATABASE_URL          string
	REDIS_URL             string
	LOG_LEVEL             string
	APP_ENV               string
	WORKER_CONCURRENCY    int
	ACCEPT_TIME_SECONDS   int
	PREP_TIME_SECONDS     int
	DELIVERY_TIME_SECONDS int
}

// LoadDotEnv carrega o .env se existir e monta a configuração a partir do ambiente
func LoadDotEnv() (*Config, error) {
	// em produção as variáveis vêm direto do ambiente; o .env é só conveniência local
	_ = godotenv.Load()

	cfg := &Config{
		PORT:         os.Getenv("PORT"),
		DATABASE_URL: os.Getenv("DATABASE_URL"),
		REDIS_URL:    os.Getenv("REDIS_URL"),
		LOG_LEVEL:    os.Getenv("LOG_LEVEL"),
		APP_ENV:      os.Getenv("APP_ENV"),
	}

	if cfg.PORT == "" {
		cfg.PORT = "8081"
	}
	if cfg.LOG_LEVEL == "" {
		cfg.LOG_LEVEL = "info"
	}
	if cfg.APP_ENV == "" {
		cfg.APP_ENV = "development"
	}

	cfg.WORKER_CONCURRENCY = intFromEnv("WORKER_CONCURRENCY", 4)
	cfg.ACCEPT_TIME_SECONDS = intFromEnv("ACCEPT_TIME_SECONDS", 2)
	cfg.PREP_TIME_SECONDS = intFromEnv("PREP_TIME_SECONDS", 5)
	cfg.DELIVERY_TIME_SECONDS = intFromEnv("DELIVERY_TIME_SECONDS", 3)

	if cfg.DATABASE_URL == "" {
		return nil, fmt.Errorf("variável de ambiente obrigatória DATABASE_URL não definida")
	}
	if cfg.REDIS_URL == "" {
		return nil, fmt.Errorf("variável de ambiente obrigatória REDIS_URL não definida")
	}

	return cfg, nil
}

func intFromEnv(name string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return defaultValue
	}

	return value
}
