package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
)

type slotRepository interface {
	ListAvailable(
		ctx context.Context,
		roomID uuid.UUID,
		from time.Time,
		to time.Time,
	) ([]model.Slot, error)
}

type SlotService struct {
	repo slotRepository
}

func NewSlotService(repo slotRepository) *SlotService {
	return &SlotService{repo: repo}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, dateString string) ([]model.Slot, error) {
	const op = "SlotService.ListAvailable"

	date, err := time.ParseInLocation("2006-01-02", dateString, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, ErrInvalidInput)
	}

	slots, err := s.repo.ListAvailable(ctx, roomID, date, date.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return slots, nil
}
