package model

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatusID int

const (
	BookingStatusActive    BookingStatusID = 1
	BookingStatusCancelled BookingStatusID = 2
)

type Booking struct {
	ID        uuid.UUID       `db:"id"`
	SlotID    uuid.UUID       `db:"slot_id"`
	UserID    uuid.UUID       `db:"user_id"`
	StatusID  BookingStatusID `db:"status_id"`
	CreatedAt time.Time       `db:"created_at"`
}
