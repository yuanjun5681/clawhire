package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string `env:"APP_ENV"   envDefault:"dev"`
	HTTPPort int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	Mongo MongoConfig
}

type MongoConfig struct {
	URI      string `env:"MONGODB_URI,required"`
	Database string `env:"MONGODB_DATABASE,required"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	cfg.AppEnv = strings.ToLower(strings.TrimSpace(cfg.AppEnv))
	cfg.LogLevel = strings.ToLower(strings.TrimSpace(cfg.LogLevel))
	return cfg, nil
}
