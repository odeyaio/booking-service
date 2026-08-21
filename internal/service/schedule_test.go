package service

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduleService_Create(t *testing.T) {
	type contextKey struct{}

	roomID := uuid.NewV7()
	txCtx := context.WithValue(context.Background(), contextKey{}, "transaction")
	generatorErr := errors.New("cannot create slots")

	tests := []struct {
		name      string
		setupMock func(*MockScheduleRepo, *MockSlotGenerator)
		wantErr   error
	}{
		{
			name: "schedule and slots created",
			setupMock: func(repo *MockScheduleRepo, generator *MockSlotGenerator) {
				repo.EXPECT().
					Create(txCtx, mock.MatchedBy(func(schedule model.Schedule) bool {
						return schedule.ID != uuid.Nil() && schedule.RoomID == roomID
					})).
					Return(nil).
					Once()
				generator.EXPECT().
					Generate(txCtx, mock.AnythingOfType("model.Schedule"), mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "room not found",
			setupMock: func(repo *MockScheduleRepo, _ *MockSlotGenerator) {
				repo.EXPECT().
					Create(txCtx, mock.AnythingOfType("model.Schedule")).
					Return(repository.ErrNotFound).
					Once()
			},
			wantErr: ErrRoomNotFound,
		},
		{
			name: "schedule already exists",
			setupMock: func(repo *MockScheduleRepo, _ *MockSlotGenerator) {
				repo.EXPECT().
					Create(txCtx, mock.AnythingOfType("model.Schedule")).
					Return(repository.ErrAlreadyExists).
					Once()
			},
			wantErr: ErrScheduleExists,
		},
		{
			name: "slot generation failed",
			setupMock: func(repo *MockScheduleRepo, generator *MockSlotGenerator) {
				repo.EXPECT().
					Create(txCtx, mock.AnythingOfType("model.Schedule")).
					Return(nil).
					Once()
				generator.EXPECT().
					Generate(txCtx, mock.AnythingOfType("model.Schedule"), mock.Anything, mock.Anything).
					Return(generatorErr).
					Once()
			},
			wantErr: generatorErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockScheduleRepo(t)
			generator := NewMockSlotGenerator(t)
			txManager := NewMockTransactionManager(t)

			txManager.EXPECT().
				Do(mock.Anything, mock.Anything).
				RunAndReturn(func(_ context.Context, fn func(context.Context) error) error {
					return fn(txCtx)
				}).
				Once()
			tt.setupMock(repo, generator)

			svc := NewScheduleService(repo, generator, txManager)
			got, err := svc.Create(context.Background(), roomID, []int{1}, "09:00", "10:00")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, model.Schedule{}, got)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil(), got.ID)
			assert.Equal(t, roomID, got.RoomID)
			assert.Equal(t, []model.Weekday{model.Monday}, got.DaysOfWeek)
			assert.Equal(t, "09:00", got.StartTime.String())
			assert.Equal(t, "10:00", got.EndTime.String())
		})
	}
}
