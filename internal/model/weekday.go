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
	if value < int(Monday) || value > int(Sunday) {
		return 0, fmt.Errorf("invalid weekday: %d", value)
	}

	return Weekday(value), nil
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

func WeekdayFromTime(day time.Weekday) (Weekday, error) {
	if day < time.Sunday || day > time.Saturday {
		return 0, fmt.Errorf("invalid time.Weekday: %d", day)
	}

	if day == time.Sunday {
		return Sunday, nil
	}

	return Weekday(day), nil
}
