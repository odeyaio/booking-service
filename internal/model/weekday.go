package model

import (
	"fmt"
	"time"
)

type Weekday uint8

const (
	Monday Weekday = iota + 1
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

func NewWeekday(value int) (Weekday, error) {
	day := Weekday(value)
	if !day.Valid() {
		return 0, fmt.Errorf("invalid weekday: %d", value)
	}

	return day, nil
}

func (d Weekday) Valid() bool {
	return d >= Monday && d <= Sunday
}

func (d Weekday) TimeWeekday() time.Weekday {
	if d == Sunday {
		return time.Sunday
	}

	return time.Weekday(d)
}

func WeekdayFromTime(day time.Weekday) Weekday {
	if day == time.Sunday {
		return Sunday
	}

	return Weekday(day)
}
