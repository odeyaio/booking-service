package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type Config struct {
	Env      Env `yaml:"env" env-default:"local"`
	Database DatabaseConfig
	Auth     AuthConfig    `yaml:"auth"`
	HTTP     HTTPConfig    `yaml:"http"`
	Workers  WorkersConfig `yaml:"workers"`
}

type DatabaseConfig struct {
	User     string `env:"DATABASE_USER" env-required:"true"`
	Password string `env:"DATABASE_PASSWORD" env-required:"true"`
	Host     string `env:"DATABASE_HOST" env-required:"true"`
	Port     uint16 `env:"DATABASE_PORT" env-default:"5432"`
	Name     string `env:"DATABASE_NAME" env-required:"true"`
}

type AuthConfig struct {
	JWTSecret  string           `env:"JWT_SECRET" env-required:"true"`
	DummyLogin DummyLoginConfig `yaml:"dummy_login"`
}

type DummyLoginConfig struct {
	Enabled     bool      `yaml:"enabled"`
	AdminUserID uuid.UUID `yaml:"admin_user_id"`
	UserUserID  uuid.UUID `yaml:"user_user_id"`
}

type HTTPConfig struct {
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"15s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"30s"`
}

type WorkersConfig struct {
	SlotGenerator SlotGeneratorWorkerConfig `yaml:"slot_generator"`
}

type SlotGeneratorWorkerConfig struct {
	Interval time.Duration `yaml:"interval" env-default:"6h"`
	Window   time.Duration `yaml:"window" env-default:"336h"`
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
	case EnvLocal, EnvDev, EnvProd:
	default:
		return fmt.Errorf("invalid environment: %q", c.Env)
	}
	if c.Workers.SlotGenerator.Interval <= 0 {
		return errors.New("workers.slot_generator.interval must be positive")
	}
	if c.Workers.SlotGenerator.Window <= 0 {
		return errors.New("workers.slot_generator.window must be positive")
	}

	if !c.Auth.DummyLogin.Enabled {
		return nil
	}
	if c.Auth.DummyLogin.AdminUserID == uuid.Nil {
		return errors.New("auth.dummy_login.admin_user_id is required when dummy login is enabled")
	}
	if c.Auth.DummyLogin.UserUserID == uuid.Nil {
		return errors.New("auth.dummy_login.user_user_id is required when dummy login is enabled")
	}

	return nil
}
