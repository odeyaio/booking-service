package config

import (
	"fmt"
	"os"
	"time"

	"github.com/odeyaio/booking-service/internal/logger"
)

type Config struct {
	Env                 logger.Env
	Addr                string
	DatabaseURL         string
	JWTSecret           string
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	HTTPShutdownTimeout time.Duration
}

func Load() (Config, error) {
	env, err := requiredEnv("APP_ENV")
	if err != nil {
		return Config{}, err
	}

	addr, err := requiredEnv("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}

	jwtSecret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}

	httpReadTimeout, err := requiredDuration("HTTP_READ_TIMEOUT")
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := requiredDuration("HTTP_WRITE_TIMEOUT")
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := requiredDuration("HTTP_IDLE_TIMEOUT")
	if err != nil {
		return Config{}, err
	}

	httpShutdownTimeout, err := requiredDuration("HTTP_SHUTDOWN_TIMEOUT")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Env:                 logger.Env(env),
		Addr:                addr,
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		HTTPReadTimeout:     httpReadTimeout,
		HTTPWriteTimeout:    httpWriteTimeout,
		HTTPIdleTimeout:     httpIdleTimeout,
		HTTPShutdownTimeout: httpShutdownTimeout,
	}, nil
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing required environment variable %s", key)
	}
	return value, nil
}

func requiredDuration(key string) (time.Duration, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration in %s: %w", key, err)
	}

	return duration, nil
}
