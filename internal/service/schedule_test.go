package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduleService_MapsRepositoryErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{
			name:    "room not found",
			repoErr: repository.ErrNotFound,
			wantErr: ErrRoomNotFound,
		},
		{
			name:    "schedule already exists",
			repoErr: repository.ErrAlreadyExists,
			wantErr: ErrScheduleExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockScheduleRepo(t)
			repo.EXPECT().
				Create(mock.Anything, mock.MatchedBy(func(schedule model.Schedule) bool {
					return schedule.RoomID != uuid.Nil
				})).
				Return(tt.repoErr).
				Once()
			svc := NewScheduleService(repo, NewMockSlotGenerator(t))

			_, err := svc.Create(
				context.Background(),
				uuid.New(),
				[]int{1},
				"09:00",
				"10:00",
			)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
