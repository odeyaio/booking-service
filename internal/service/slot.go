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

type SlotService struct {
	repo     slotRepository
	roomRepo slotRoomRepository
}

func NewSlotService(repo slotRepository, roomRepo slotRoomRepository) *SlotService {
	return &SlotService{repo: repo, roomRepo: roomRepo}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, dateString string) ([]model.Slot, error) {
	const op = "SlotService.ListAvailable"

	date, err := time.ParseInLocation("2006-01-02", dateString, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, ErrInvalidInput)
	}

	if _, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrRoomNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slots, err := s.repo.ListAvailable(ctx, roomID, date, date.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return slots, nil
}
