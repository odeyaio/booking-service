package logger

import (
	"fmt"
	"log/slog"
	"os"
)

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

func Setup(env Env) (*slog.Logger, error) {
	switch env {
	case EnvLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), nil
	case EnvDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), nil
	case EnvProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})), nil
	default:
		return nil, fmt.Errorf("unsupported env %q", env)
	}
}
