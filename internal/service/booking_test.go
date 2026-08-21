package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/pagination"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBookingService_Create(t *testing.T) {
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	createdAt := now.Add(time.Minute)
	userID := uuid.New()
	slotID := uuid.New()

	tests := []struct {
		name       string
		slot       model.Slot
		slotErr    error
		createErr  error
		wantErr    error
		wantCreate bool
	}{
		{
			name:       "booking created",
			slot:       model.Slot{ID: slotID, Start: now.Add(time.Hour)},
			wantCreate: true,
		},
		{
			name:    "slot not found",
			slotErr: repository.ErrNotFound,
			wantErr: ErrSlotNotFound,
		},
		{
			name:    "past slot",
			slot:    model.Slot{ID: slotID, Start: now.Add(-time.Second)},
			wantErr: ErrInvalidInput,
		},
		{
			name:       "slot already booked",
			slot:       model.Slot{ID: slotID, Start: now.Add(time.Hour)},
			createErr:  repository.ErrAlreadyExists,
			wantErr:    ErrSlotAlreadyBooked,
			wantCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBookingRepository(t)
			slotRepo := NewMockBookingSlotRepository(t)
			slotRepo.EXPECT().
				GetByID(mock.Anything, slotID).
				Return(tt.slot, tt.slotErr).
				Once()
			if tt.wantCreate {
				repo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Booking")).
					Run(func(_ context.Context, booking *model.Booking) {
						assert.NotEqual(t, uuid.Nil, booking.ID)
						assert.Equal(t, slotID, booking.SlotID)
						assert.Equal(t, userID, booking.UserID)
						assert.Equal(t, model.BookingStatusActive, booking.StatusID)
						booking.CreatedAt = createdAt
					}).
					Return(tt.createErr).
					Once()
			}

			svc := NewBookingService(repo, slotRepo)
			svc.now = func() time.Time { return now }

			booking, err := svc.Create(context.Background(), userID, slotID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, createdAt, booking.CreatedAt)
		})
	}
}

func TestBookingService_List(t *testing.T) {
	repo := NewMockBookingRepository(t)
	repo.EXPECT().
		List(mock.Anything, 20, 40).
		Return([]model.Booking{{ID: uuid.New()}}, 45, nil).
		Once()
	svc := NewBookingService(repo, NewMockBookingSlotRepository(t))

	bookings, total, err := svc.List(context.Background(), pagination.Params{Page: 3, PageSize: 20})
	require.NoError(t, err)
	assert.Len(t, bookings, 1)
	assert.Equal(t, 45, total)

	for _, args := range [][2]int{{0, 20}, {1, 0}, {1, 101}} {
		_, _, err = svc.List(context.Background(), pagination.Params{Page: args[0], PageSize: args[1]})
		require.ErrorIs(t, err, ErrInvalidInput)
	}
}

func TestBookingService_ListMy(t *testing.T) {
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	want := []model.Booking{{ID: uuid.New(), UserID: userID}}
	repo := NewMockBookingRepository(t)
	repo.EXPECT().
		ListByUser(mock.Anything, userID, now).
		Return(want, nil).
		Once()
	svc := NewBookingService(repo, NewMockBookingSlotRepository(t))
	svc.now = func() time.Time { return now }

	got, err := svc.ListMy(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestBookingService_Cancel(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name       string
		booking    model.Booking
		getErr     error
		cancelErr  error
		wantErr    error
		wantCancel bool
	}{
		{
			name:       "booking cancelled",
			booking:    model.Booking{ID: bookingID, UserID: userID},
			wantCancel: true,
		},
		{
			name:    "booking not found",
			getErr:  repository.ErrNotFound,
			wantErr: ErrBookingNotFound,
		},
		{
			name:    "another user's booking",
			booking: model.Booking{ID: bookingID, UserID: otherUserID},
			wantErr: ErrForbidden,
		},
		{
			name:       "booking disappeared before update",
			booking:    model.Booking{ID: bookingID, UserID: userID},
			cancelErr:  repository.ErrNotFound,
			wantErr:    ErrBookingNotFound,
			wantCancel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockBookingRepository(t)
			repo.EXPECT().
				GetByID(mock.Anything, bookingID).
				Return(tt.booking, tt.getErr).
				Once()
			if tt.wantCancel {
				repo.EXPECT().
					Cancel(mock.Anything, bookingID).
					RunAndReturn(func(context.Context, uuid.UUID) (model.Booking, error) {
						booking := tt.booking
						booking.StatusID = model.BookingStatusCancelled
						return booking, tt.cancelErr
					}).
					Once()
			}
			svc := NewBookingService(repo, NewMockBookingSlotRepository(t))

			booking, err := svc.Cancel(context.Background(), userID, bookingID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, model.BookingStatusCancelled, booking.StatusID)
		})
	}
}

func TestBookingService_RejectsEmptyIDs(t *testing.T) {
	svc := NewBookingService(NewMockBookingRepository(t), NewMockBookingSlotRepository(t))

	_, err := svc.Create(context.Background(), uuid.Nil(), uuid.New())
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = svc.ListMy(context.Background(), uuid.Nil())
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = svc.Cancel(context.Background(), uuid.New(), uuid.Nil())
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestBookingService_PropagatesUnexpectedError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repo := NewMockBookingRepository(t)
	repo.EXPECT().
		List(mock.Anything, 20, 0).
		Return(nil, 0, wantErr).
		Once()
	svc := NewBookingService(repo, NewMockBookingSlotRepository(t))

	_, _, err := svc.List(context.Background(), pagination.Params{Page: 1, PageSize: 20})
	require.ErrorIs(t, err, wantErr)
}
