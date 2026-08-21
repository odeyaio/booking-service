package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
)

type scheduleRepo interface {
	Create(ctx context.Context, schedule model.Schedule) error
}

type slotGenerator interface {
	Generate(ctx context.Context, schedule model.Schedule, from, to time.Time) error
}

type transactionManager interface {
	Do(ctx context.Context, fn func(context.Context) error) error
}

type ScheduleService struct {
	scheduleRepo  scheduleRepo
	slotGenerator slotGenerator
	txManager     transactionManager
}

func NewScheduleService(
	scheduleRepo scheduleRepo,
	slotGenerator slotGenerator,
	txManager transactionManager,
) *ScheduleService {
	return &ScheduleService{
		scheduleRepo:  scheduleRepo,
		slotGenerator: slotGenerator,
		txManager:     txManager,
	}
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
		ID:         uuid.NewV7(),
		RoomID:     roomID,
		DaysOfWeek: weekdays,
		StartTime:  startTimeOfDay,
		EndTime:    endTimeOfDay,
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

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
			return err
		}

		if err := s.slotGenerator.Generate(ctx, schedule, from, to); err != nil {
			return fmt.Errorf("generate slots: %w", err)
		}

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			err = ErrRoomNotFound
		case errors.Is(err, repository.ErrAlreadyExists):
			err = ErrScheduleExists
		}
		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	return schedule, nil
}
