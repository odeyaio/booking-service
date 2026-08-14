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
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	date := time.Date(2026, time.August, 16, 0, 0, 0, 0, time.UTC)
	bookingHorizon := 90 * 24 * time.Hour
	wantSlots := []model.Slot{{ID: uuid.New(), RoomID: roomID}}
	schedule := model.Schedule{ID: uuid.New(), RoomID: roomID}

	newService := func(t *testing.T) (
		*SlotService,
		*MockSlotRepository,
		*MockSlotRoomRepository,
		*MockSlotScheduleRepository,
		*MockSlotGenerator,
	) {
		t.Helper()

		slotRepo := NewMockSlotRepository(t)
		roomRepo := NewMockSlotRoomRepository(t)
		scheduleRepo := NewMockSlotScheduleRepository(t)
		generator := NewMockSlotGenerator(t)
		svc := NewSlotService(slotRepo, roomRepo, scheduleRepo, generator, bookingHorizon)
		svc.now = func() time.Time { return now }

		return svc, slotRepo, roomRepo, scheduleRepo, generator
	}

	t.Run("room not found", func(t *testing.T) {
		svc, _, roomRepo, _, _ := newService(t)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{}, repository.ErrNotFound).
			Once()

		_, err := svc.ListAvailable(context.Background(), roomID, "2026-08-16")

		require.ErrorIs(t, err, ErrRoomNotFound)
	})

	t.Run("available slots", func(t *testing.T) {
		svc, slotRepo, roomRepo, scheduleRepo, generator := newService(t)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{ID: roomID}, nil).
			Once()
		scheduleRepo.EXPECT().
			GetByRoomID(mock.Anything, roomID).
			Return(schedule, nil).
			Once()
		generator.EXPECT().
			Generate(mock.Anything, schedule, date, date.AddDate(0, 0, 1)).
			Return(nil).
			Once()
		slotRepo.EXPECT().
			ListAvailable(mock.Anything, roomID, date, date.AddDate(0, 0, 1)).
			Return(wantSlots, nil).
			Once()

		slots, err := svc.ListAvailable(context.Background(), roomID, "2026-08-16")

		require.NoError(t, err)
		assert.Equal(t, wantSlots, slots)
	})

	t.Run("today only includes future slots", func(t *testing.T) {
		svc, slotRepo, roomRepo, scheduleRepo, generator := newService(t)
		today := time.Date(2026, time.August, 15, 0, 0, 0, 0, time.UTC)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{ID: roomID}, nil).
			Once()
		scheduleRepo.EXPECT().
			GetByRoomID(mock.Anything, roomID).
			Return(schedule, nil).
			Once()
		generator.EXPECT().
			Generate(mock.Anything, schedule, now, today.AddDate(0, 0, 1)).
			Return(nil).
			Once()
		slotRepo.EXPECT().
			ListAvailable(mock.Anything, roomID, now, today.AddDate(0, 0, 1)).
			Return(wantSlots, nil).
			Once()

		slots, err := svc.ListAvailable(context.Background(), roomID, "2026-08-15")

		require.NoError(t, err)
		assert.Equal(t, wantSlots, slots)
	})

	t.Run("room without schedule", func(t *testing.T) {
		svc, _, roomRepo, scheduleRepo, _ := newService(t)
		roomRepo.EXPECT().
			GetByID(mock.Anything, roomID).
			Return(model.Room{ID: roomID}, nil).
			Once()
		scheduleRepo.EXPECT().
			GetByRoomID(mock.Anything, roomID).
			Return(model.Schedule{}, repository.ErrNotFound).
			Once()

		slots, err := svc.ListAvailable(context.Background(), roomID, "2026-08-16")

		require.NoError(t, err)
		assert.Empty(t, slots)
	})

	t.Run("invalid date", func(t *testing.T) {
		svc, _, _, _, _ := newService(t)

		_, err := svc.ListAvailable(context.Background(), roomID, "15-08-2026")

		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("past date", func(t *testing.T) {
		svc, _, _, _, _ := newService(t)

		_, err := svc.ListAvailable(context.Background(), roomID, "2026-08-14")

		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("date outside booking horizon", func(t *testing.T) {
		svc, _, _, _, _ := newService(t)

		_, err := svc.ListAvailable(context.Background(), roomID, "2026-11-14")

		require.ErrorIs(t, err, ErrInvalidInput)
	})
}
