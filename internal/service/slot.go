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

type slotRepository interface {
	ListAvailable(
		ctx context.Context,
		roomID uuid.UUID,
		from time.Time,
		to time.Time,
	) ([]model.Slot, error)
}

type slotRoomRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (model.Room, error)
}

type slotScheduleRepository interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (model.Schedule, error)
}

type SlotService struct {
	repo           slotRepository
	roomRepo       slotRoomRepository
	scheduleRepo   slotScheduleRepository
	slotGenerator  slotGenerator
	bookingHorizon time.Duration
	now            func() time.Time
}

func NewSlotService(
	repo slotRepository,
	roomRepo slotRoomRepository,
	scheduleRepo slotScheduleRepository,
	slotGenerator slotGenerator,
	bookingHorizon time.Duration,
) *SlotService {
	return &SlotService{
		repo:           repo,
		roomRepo:       roomRepo,
		scheduleRepo:   scheduleRepo,
		slotGenerator:  slotGenerator,
		bookingHorizon: bookingHorizon,
		now:            time.Now,
	}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, dateString string) ([]model.Slot, error) {
	const op = "SlotService.ListAvailable"

	date, err := time.ParseInLocation("2006-01-02", dateString, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, ErrInvalidInput)
	}
	now := s.now().UTC()
	today := startOfDay(now)
	if date.Before(today) || date.After(startOfDay(now.Add(s.bookingHorizon))) {
		return nil, fmt.Errorf("%s: date is outside booking horizon: %w", op, ErrInvalidInput)
	}
	from := date
	if date.Equal(today) {
		from = now
	}
	to := date.AddDate(0, 0, 1)

	if _, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrRoomNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	schedule, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return []model.Slot{}, nil
		}
		return nil, fmt.Errorf("%s: get schedule: %w", op, err)
	}

	if err := s.slotGenerator.Generate(ctx, schedule, from, to); err != nil {
		return nil, fmt.Errorf("%s: generate slots: %w", op, err)
	}

	slots, err := s.repo.ListAvailable(ctx, roomID, from, to)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return slots, nil
}
