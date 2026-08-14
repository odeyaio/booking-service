package model

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID         uuid.UUID `db:"id"`
	RoomID     uuid.UUID `db:"room_id"`
	ScheduleID uuid.UUID `db:"schedule_id"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
}
