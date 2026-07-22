package room

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	description := "Meeting room"
	capacity := 10
	zeroCapacity := 0
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		roomName    string
		description *string
		capacity    *int
		setupMock   func(*MockRoomRepository)
		want        *model.Room
		wantErr     error
	}{
		{
			name:        "room created",
			roomName:    "  Room 1  ",
			description: &description,
			capacity:    &capacity,
			setupMock: func(repo *MockRoomRepository) {
				repo.EXPECT().
					Create(ctx, mock.MatchedBy(func(room *model.Room) bool {
						return room.ID != uuid.Nil &&
							room.Name == "Room 1" &&
							room.Description == &description &&
							room.Capacity == &capacity
					})).
					Return(nil).
					Once()
			},
			want: &model.Room{
				Name:        "Room 1",
				Description: &description,
				Capacity:    &capacity,
			},
		},
		{
			name:     "empty name",
			roomName: "   ",
			wantErr:  service.ErrInvalidInput,
		},
		{
			name:     "zero capacity",
			roomName: "Room 1",
			capacity: &zeroCapacity,
			wantErr:  service.ErrInvalidInput,
		},
		{
			name:     "repository error",
			roomName: "Room 1",
			setupMock: func(repo *MockRoomRepository) {
				repo.EXPECT().
					Create(ctx, mock.MatchedBy(func(room *model.Room) bool {
						return room.ID != uuid.Nil && room.Name == "Room 1"
					})).
					Return(repoErr).
					Once()
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRoomRepository(t)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			got, err := New(repo).Create(
				ctx,
				tt.roomName,
				tt.description,
				tt.capacity,
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, model.Room{}, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.want)
			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.Capacity, got.Capacity)
		})
	}
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("repository error")
	rooms := []model.Room{
		{
			ID:   uuid.MustParse("12341234-1234-1234-1234-123412341234"),
			Name: "Room 1",
		},
	}

	tests := []struct {
		name      string
		setupMock func(*MockRoomRepository)
		want      []model.Room
		wantErr   error
	}{
		{
			name: "rooms returned",
			setupMock: func(repo *MockRoomRepository) {
				repo.EXPECT().
					List(ctx).
					Return(rooms, nil).
					Once()
			},
			want: rooms,
		},
		{
			name: "empty list returned",
			setupMock: func(repo *MockRoomRepository) {
				repo.EXPECT().
					List(ctx).
					Return([]model.Room{}, nil).
					Once()
			},
			want: []model.Room{},
		},
		{
			name: "repository error",
			setupMock: func(repo *MockRoomRepository) {
				repo.EXPECT().
					List(ctx).
					Return(nil, repoErr).
					Once()
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRoomRepository(t)
			tt.setupMock(repo)

			got, err := New(repo).List(ctx)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
