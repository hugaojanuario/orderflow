package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config espelha os nomes das variáveis de ambiente em maiúsculas
type Config struct {
	PORT         string
	DATABASE_URL string
	REDIS_URL    string
	JWT_SECRET   string
	LOG_LEVEL    string
	APP_ENV      string
}

// LoadDotEnv carrega o .env se existir e monta a configuração a partir do ambiente
func LoadDotEnv() (*Config, error) {
	// em produção as variáveis vêm direto do ambiente; o .env é só conveniência local
	_ = godotenv.Load()

	cfg := &Config{
		PORT:         os.Getenv("PORT"),
		DATABASE_URL: os.Getenv("DATABASE_URL"),
		REDIS_URL:    os.Getenv("REDIS_URL"),
		JWT_SECRET:   os.Getenv("JWT_SECRET"),
		LOG_LEVEL:    os.Getenv("LOG_LEVEL"),
		APP_ENV:      os.Getenv("APP_ENV"),
	}

	if cfg.PORT == "" {
		cfg.PORT = "8080"
	}
	if cfg.LOG_LEVEL == "" {
		cfg.LOG_LEVEL = "info"
	}
	if cfg.APP_ENV == "" {
		cfg.APP_ENV = "development"
	}

	if cfg.DATABASE_URL == "" {
		return nil, fmt.Errorf("variável de ambiente obrigatória DATABASE_URL não definida")
	}
	if cfg.REDIS_URL == "" {
		return nil, fmt.Errorf("variável de ambiente obrigatória REDIS_URL não definida")
	}
	if cfg.JWT_SECRET == "" {
		return nil, fmt.Errorf("variável de ambiente obrigatória JWT_SECRET não definida")
	}

	return cfg, nil
}
