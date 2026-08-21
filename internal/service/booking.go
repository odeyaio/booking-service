package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/pagination"
	"github.com/odeyaio/booking-service/internal/repository"
)

type bookingRepository interface {
	Create(ctx context.Context, booking *model.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (model.Booking, error)
	List(ctx context.Context, limit, offset int) ([]model.Booking, int, error)
	ListByUser(ctx context.Context, userID uuid.UUID, from time.Time) ([]model.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) (model.Booking, error)
}

type bookingSlotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (model.Slot, error)
}

type BookingService struct {
	repo     bookingRepository
	slotRepo bookingSlotRepository
	now      func() time.Time
}

func NewBookingService(repo bookingRepository, slotRepo bookingSlotRepository) *BookingService {
	return &BookingService{
		repo:     repo,
		slotRepo: slotRepo,
		now:      time.Now,
	}
}

func (s *BookingService) Create(
	ctx context.Context,
	userID uuid.UUID,
	slotID uuid.UUID,
) (model.Booking, error) {
	const op = "BookingService.Create"

	if userID == uuid.Nil() || slotID == uuid.Nil() {
		return model.Booking{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrSlotNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}
	if slot.Start.Before(s.now().UTC()) {
		return model.Booking{}, fmt.Errorf("%s: slot is in the past: %w", op, ErrInvalidInput)
	}

	booking := model.Booking{
		ID:       uuid.New(),
		SlotID:   slotID,
		UserID:   userID,
		StatusID: model.BookingStatusActive,
	}
	if err := s.repo.Create(ctx, &booking); err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyExists):
			err = ErrSlotAlreadyBooked
		case errors.Is(err, repository.ErrNotFound):
			err = ErrSlotNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}

func (s *BookingService) List(
	ctx context.Context,
	params pagination.Params,
) ([]model.Booking, int, error) {
	const op = "BookingService.List"

	if err := params.Validate(); err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	bookings, total, err := s.repo.List(ctx, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, total, nil
}

func (s *BookingService) ListMy(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	const op = "BookingService.ListMy"

	if userID == uuid.Nil() {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	bookings, err := s.repo.ListByUser(ctx, userID, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, nil
}

func (s *BookingService) Cancel(
	ctx context.Context,
	userID uuid.UUID,
	bookingID uuid.UUID,
) (model.Booking, error) {
	const op = "BookingService.Cancel"

	if userID == uuid.Nil() || bookingID == uuid.Nil() {
		return model.Booking{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrBookingNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}
	if booking.UserID != userID {
		return model.Booking{}, fmt.Errorf("%s: cannot cancel another user's booking: %w", op, ErrForbidden)
	}

	booking, err = s.repo.Cancel(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrBookingNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}
