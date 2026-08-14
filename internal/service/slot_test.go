package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlotService_ListAvailable(t *testing.T) {
	roomID := uuid.New()
	date := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
	wantSlots := []model.Slot{{ID: uuid.New(), RoomID: roomID}}

	t.Run("room not found", func(t *testing.T) {
		slotRepo := NewMockSlotRepository(t)
		roomRepo := NewMockSlotRoomRepository(t)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{}, repository.ErrNotFound).
			Once()
		svc := NewSlotService(slotRepo, roomRepo)

		_, err := svc.ListAvailable(context.Background(), roomID, "2026-08-15")

		require.ErrorIs(t, err, ErrRoomNotFound)
	})

	t.Run("available slots", func(t *testing.T) {
		slotRepo := NewMockSlotRepository(t)
		roomRepo := NewMockSlotRoomRepository(t)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{ID: roomID}, nil).
			Once()
		slotRepo.EXPECT().
			ListAvailable(mock.Anything, roomID, date, date.AddDate(0, 0, 1)).
			Return(wantSlots, nil).
			Once()
		svc := NewSlotService(slotRepo, roomRepo)

		slots, err := svc.ListAvailable(context.Background(), roomID, "2026-08-15")

		require.NoError(t, err)
		assert.Equal(t, wantSlots, slots)
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := NewSlotService(NewMockSlotRepository(t), NewMockSlotRoomRepository(t))

		_, err := svc.ListAvailable(context.Background(), roomID, "15-08-2026")

		require.ErrorIs(t, err, ErrInvalidInput)
	})
}
