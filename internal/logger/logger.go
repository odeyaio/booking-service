package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/odeyaio/booking-service/internal/config"
)

func Setup(env config.Env) (*slog.Logger, error) {
	switch env {
	case config.EnvLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), nil
	case config.EnvDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), nil
	case config.EnvProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})), nil
	default:
		return nil, fmt.Errorf("unsupported env %q", env)
	}
}
