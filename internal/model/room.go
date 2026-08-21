package model

import (
	"time"
	"uuid"
)

type Room struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Capacity    *int      `db:"capacity"`
	CreatedAt   time.Time `db:"created_at"`
}
