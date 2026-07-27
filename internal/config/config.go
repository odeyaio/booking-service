package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/odeyaio/booking-service/internal/logger"
)

type Config struct {
	Env      logger.Env `yaml:"env" env-default:"local"`
	Database DatabaseConfig
	JWT      JWTConfig
	HTTP     HTTPConfig `yaml:"http"`
}

type DatabaseConfig struct {
	User     string `env:"DATABASE_USER" env-required:"true"`
	Password string `env:"DATABASE_PASSWORD" env-required:"true"`
	Host     string `env:"DATABASE_HOST" env-required:"true"`
	Port     uint16 `env:"DATABASE_PORT" env-default:"5432"`
	Name     string `env:"DATABASE_NAME" env-required:"true"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" env-required:"true"`
}

type HTTPConfig struct {
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"15s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"30s"`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		return Config{}, errors.New("CONFIG_PATH is not set")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		return Config{}, fmt.Errorf("load %s: %w", cfgPath, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Env {
	case logger.EnvLocal, logger.EnvDev, logger.EnvProd:
		return nil
	default:
		return fmt.Errorf("invalid environment: %q", c.Env)
	}
}
