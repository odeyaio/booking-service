package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/odeyaio/booking-service/internal/logger"
)

type Config struct {
	Env      logger.Env
	Addr     string
	Database DatabaseConfig
	JWT      JWTConfig
	HTTP     HTTPConfig
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
}

type HTTPConfig struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	l := envLoader{}

	cfg := Config{
		Env:  l.env("APP_ENV"),
		Addr: l.string("HTTP_ADDR"),
		Database: DatabaseConfig{
			URL: l.string("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret: l.string("JWT_SECRET"),
		},
		HTTP: HTTPConfig{
			ReadTimeout:     l.duration("HTTP_READ_TIMEOUT"),
			WriteTimeout:    l.duration("HTTP_WRITE_TIMEOUT"),
			IdleTimeout:     l.duration("HTTP_IDLE_TIMEOUT"),
			ShutdownTimeout: l.duration("HTTP_SHUTDOWN_TIMEOUT"),
		},
	}

	if l.err != nil {
		return Config{}, l.err
	}

	return cfg, nil
}

type envLoader struct {
	err error
}

func (l *envLoader) string(key string) string {
	if l.err != nil {
		return ""
	}
	value := os.Getenv(key)
	if value == "" {
		l.err = fmt.Errorf("missing required environment variable %s", key)
		return ""
	}
	return value
}

func (l *envLoader) env(key string) logger.Env {
	value := logger.Env(l.string(key))
	if l.err != nil {
		return ""
	}

	switch value {
	case logger.EnvLocal, logger.EnvDev, logger.EnvProd:
		return value
	default:
		l.err = fmt.Errorf("invalid environment in %s: %s", key, value)
		return ""
	}
}

func (l *envLoader) duration(key string) time.Duration {
	value := l.string(key)
	if l.err != nil {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		l.err = fmt.Errorf("invalid duration in %s: %w", key, err)
		return 0
	}
	return duration
}
