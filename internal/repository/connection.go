package repository

import (
	"fmt"

	"github.com/odeyaio/booking-service/internal/config"
)

func ConnectionURL(cfg config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)
}
