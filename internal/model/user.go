package model

import (
	"time"

	"github.com/google/uuid"
)

type RoleID int

const (
	RoleAdmin RoleID = 1
	RoleUser  RoleID = 2
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	RoleID       RoleID    `db:"role_id"`
	CreatedAt    time.Time `db:"created_at"`
}
