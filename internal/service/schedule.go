package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
)

type scheduleRepo interface {
	Create(ctx context.Context, schedule model.Schedule) error
}

type slotGenerator interface {
	Generate(ctx context.Context, schedule model.Schedule, from, to time.Time) error
}

type ScheduleService struct {
	scheduleRepo  scheduleRepo
	slotGenerator slotGenerator
}

func NewScheduleService(scheduleRepo scheduleRepo, slotGenerator slotGenerator) *ScheduleService {
	return &ScheduleService{scheduleRepo: scheduleRepo, slotGenerator: slotGenerator}
}

func (s *ScheduleService) Create(
	ctx context.Context,
	roomID uuid.UUID,
	daysOfWeek []int,
	startTime string,
	endTime string) (model.Schedule, error) {
	const op = "ScheduleService.Create"

	weekdays := make([]model.Weekday, 0, len(daysOfWeek))
	for _, day := range daysOfWeek {
		weekday, err := model.NewWeekday(day)
		if err != nil {
			return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
		}

		weekdays = append(weekdays, weekday)
	}

	startTimeOfDay, err := model.ParseTimeOfDay(startTime)
	if err != nil {
		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}
	endTimeOfDay, err := model.ParseTimeOfDay(endTime)
	if err != nil {
		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	if startTimeOfDay >= endTimeOfDay {
		return model.Schedule{}, fmt.Errorf("%s: start time must precede end time: %w", op, ErrInvalidInput)
	}

	schedule := model.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: weekdays,
		StartTime:  startTimeOfDay,
		EndTime:    endTimeOfDay,
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			err = ErrRoomNotFound
		case errors.Is(err, repository.ErrAlreadyExists):
			err = ErrScheduleExists
		}
		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now().UTC()
	from := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.UTC,
	)
	to := from.AddDate(0, 0, 14)

	if err := s.slotGenerator.Generate(ctx, schedule, from, to); err != nil {
		return model.Schedule{}, fmt.Errorf("%s: generate slots: %w", op, err)
	}

	return schedule, nil
}
