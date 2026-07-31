package model

import (
	"fmt"
	"time"
)

type TimeOfDay uint16

const minutesPerDay = 24 * 60

func NewTimeOfDay(hour, minute int) (TimeOfDay, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("invalid time of day: %02d:%02d", hour, minute)
	}

	return TimeOfDay(hour*60 + minute), nil
}

func ParseTimeOfDay(value string) (TimeOfDay, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, err
	}

	return NewTimeOfDay(parsed.Hour(), parsed.Minute())
}

func (t TimeOfDay) Hour() int {
	return int(t) / 60
}

func (t TimeOfDay) Minute() int {
	return int(t) % 60
}

func (t TimeOfDay) Valid() bool {
	return t < minutesPerDay
}

func (t TimeOfDay) At(date time.Time) time.Time {
	date = date.UTC()

	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		t.Hour(),
		t.Minute(),
		0,
		0,
		time.UTC,
	)
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}
