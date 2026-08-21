package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlotGeneratorWorker_RunOnce(t *testing.T) {
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	window := 14 * 24 * time.Hour
	schedules := []model.Schedule{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}
	listErr := errors.New("cannot list schedules")
	generatorErr := errors.New("database unavailable")

	tests := []struct {
		name      string
		listErr   error
		setupMock func(*MockSlotGenerator)
		wantErr   error
	}{
		{
			name: "slots generated for all schedules",
			setupMock: func(generator *MockSlotGenerator) {
				for _, schedule := range schedules {
					generator.EXPECT().
						Generate(mock.Anything, schedule, now, now.Add(window)).
						Return(nil).
						Once()
				}
			},
		},
		{
			name:    "schedule list failed",
			listErr: listErr,
			wantErr: listErr,
		},
		{
			name: "generation continues after schedule failure",
			setupMock: func(generator *MockSlotGenerator) {
				generator.EXPECT().
					Generate(mock.Anything, schedules[0], now, now.Add(window)).
					Return(generatorErr).
					Once()
				generator.EXPECT().
					Generate(mock.Anything, schedules[1], now, now.Add(window)).
					Return(nil).
					Once()
			},
			wantErr: generatorErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockScheduleRepository(t)
			generator := NewMockSlotGenerator(t)
			repo.EXPECT().
				List(mock.Anything).
				Return(schedules, tt.listErr).
				Once()
			if tt.setupMock != nil {
				tt.setupMock(generator)
			}

			worker := NewSlotGeneratorWorker(
				repo,
				generator,
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				time.Hour,
				window,
			)
			worker.now = func() time.Time { return now }

			err := worker.RunOnce(context.Background())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
