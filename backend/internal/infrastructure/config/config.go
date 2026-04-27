package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string `env:"APP_ENV"   envDefault:"dev"`
	HTTPPort int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	Mongo       MongoConfig
	Auth        AuthConfig
	ClawSynapse ClawSynapseConfig
}

type MongoConfig struct {
	URI      string `env:"MONGODB_URI,required"`
	Database string `env:"MONGODB_DATABASE,required"`
}

// AuthConfig 控制 human 账号的注册/登录与 JWT 签发。
// JWTSecret 在生产环境应通过环境变量注入；开发态下给一个占位默认值，方便起服。
type AuthConfig struct {
	JWTSecret   string        `env:"JWT_SECRET"        envDefault:"clawhire-dev-secret-change-me"`
	JWTTTL      time.Duration `env:"JWT_TTL"           envDefault:"72h"`
	JWTIssuer   string        `env:"JWT_ISSUER"        envDefault:"clawhire"`
	BcryptCost  int           `env:"BCRYPT_COST"       envDefault:"12"`
	MinPassword int           `env:"MIN_PASSWORD_LEN"  envDefault:"8"`
}

// ClawSynapseConfig 控制出站消息发布。NodeAPIURL 为空时禁用跨平台同步。
type ClawSynapseConfig struct {
	NodeAPIURL             string `env:"CLAWSYNAPSE_NODE_API_URL"`
	DefaultTrustMeshNodeID string `env:"TRUSTMESH_PLATFORM_NODE_ID"`
	TrustMeshWebURL        string `env:"TRUSTMESH_WEB_URL"`
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
