package service

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotAlreadyBooked  = errors.New("slot already booked")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrRoomNotFound       = errors.New("room not found")
	ErrScheduleExists     = errors.New("schedule already exists")
	ErrForbidden          = errors.New("forbidden")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
