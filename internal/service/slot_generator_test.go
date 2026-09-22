package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlotGenerator_Generate(t *testing.T) {
	roomID := uuid.NewV7()
	schedule := model.Schedule{
		ID:         uuid.NewV7(),
		RoomID:     roomID,
		DaysOfWeek: []model.Weekday{model.Monday},
		StartTime:  model.TimeOfDay(9 * 60),
		EndTime:    model.TimeOfDay(10 * 60),
	}
	monday := time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC)
	from := monday
	to := monday.AddDate(0, 0, 1)

	t.Run("creates missing slots", func(t *testing.T) {
		creator := NewMockSlotCreator(t)
		generator := NewSlotGenerator(creator)
		creator.EXPECT().
			CountInRange(mock.Anything, roomID, from, to).
			Return(1, nil).
			Once()
		creator.EXPECT().
			UpsertBatch(mock.Anything, mock.Anything).
			Run(func(_ context.Context, slots []model.Slot) {
				require.Len(t, slots, 2)
				assert.Equal(t, monday.Add(9*time.Hour), slots[0].Start)
				assert.Equal(t, monday.Add(9*time.Hour+30*time.Minute), slots[1].Start)
				assert.Equal(t, monday.Add(10*time.Hour), slots[1].End)
			}).
			Return(nil).
			Once()

		err := generator.Generate(context.Background(), schedule, from, to)

		require.NoError(t, err)
	})

	t.Run("skips write when all slots exist", func(t *testing.T) {
		creator := NewMockSlotCreator(t)
		generator := NewSlotGenerator(creator)
		creator.EXPECT().
			CountInRange(mock.Anything, roomID, from, to).
			Return(2, nil).
			Once()

		err := generator.Generate(context.Background(), schedule, from, to)

		require.NoError(t, err)
	})

	t.Run("no slots expected on non-schedule day", func(t *testing.T) {
		creator := NewMockSlotCreator(t)
		generator := NewSlotGenerator(creator)
		tuesday := monday.AddDate(0, 0, 1)

		err := generator.Generate(context.Background(), schedule, tuesday, tuesday.AddDate(0, 0, 1))

		require.NoError(t, err)
	})

	t.Run("count error", func(t *testing.T) {
		creator := NewMockSlotCreator(t)
		generator := NewSlotGenerator(creator)
		countErr := errors.New("db down")
		creator.EXPECT().
			CountInRange(mock.Anything, roomID, from, to).
			Return(0, countErr).
			Once()

		err := generator.Generate(context.Background(), schedule, from, to)

		require.ErrorIs(t, err, countErr)
	})

	t.Run("invalid range", func(t *testing.T) {
		generator := NewSlotGenerator(NewMockSlotCreator(t))

		err := generator.Generate(context.Background(), schedule, to, from)

		require.ErrorIs(t, err, ErrInvalidInput)
	})
}
