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

type BookingStatusID int

const (
	BookingStatusActive    BookingStatusID = 1
	BookingStatusCancelled BookingStatusID = 2
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	RoleID       RoleID    `db:"role_id"`
	CreatedAt    time.Time `db:"created_at"`
}

type Room struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Capacity    *int      `db:"capacity"`
	CreatedAt   time.Time `db:"created_at"`
}

type Schedule struct {
	ID         uuid.UUID `db:"id"`
	RoomID     uuid.UUID `db:"room_id"`
	DaysOfWeek []int     `db:"days_of_week"`
	StartTime  string    `db:"start_time"`
	EndTime    string    `db:"end_time"`
}

type Slot struct {
	ID         uuid.UUID `db:"id"`
	RoomID     uuid.UUID `db:"room_id"`
	ScheduleID uuid.UUID `db:"schedule_id"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
}

type Booking struct {
	ID        uuid.UUID       `db:"id"`
	SlotID    uuid.UUID       `db:"slot_id"`
	UserID    uuid.UUID       `db:"user_id"`
	StatusID  BookingStatusID `db:"status_id"`
	CreatedAt time.Time       `db:"created_at"`
}
