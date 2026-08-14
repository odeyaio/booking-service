package model

import "github.com/google/uuid"

type Schedule struct {
	ID         uuid.UUID `db:"id"`
	RoomID     uuid.UUID `db:"room_id"`
	DaysOfWeek []Weekday `db:"days_of_week"`
	StartTime  TimeOfDay `db:"start_time"`
	EndTime    TimeOfDay `db:"end_time"`
}
