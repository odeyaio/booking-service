package handler

import (
	"net/http"
	"testing"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/mock"
)

func TestScheduleHandler_MapsServiceErrors(t *testing.T) {
	roomID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tests := []struct {
		name        string
		serviceErr  error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "room not found",
			serviceErr:  service.ErrRoomNotFound,
			wantStatus:  http.StatusNotFound,
			wantCode:    ErrorCodeRoomNotFound,
			wantMessage: "room not found",
		},
		{
			name:        "schedule already exists",
			serviceErr:  service.ErrScheduleExists,
			wantStatus:  http.StatusConflict,
			wantCode:    ErrorCodeScheduleExists,
			wantMessage: "schedule already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockScheduleService(t)
			svc.EXPECT().
				Create(mock.Anything, roomID, []int{1}, "09:00", "10:00").
				Return(model.Schedule{}, tt.serviceErr).
				Once()
			e := NewEcho()
			e.POST("/rooms/:roomId/schedule/create", NewScheduleHandler(svc).Create)

			rec := PerformJSONRequest(
				t,
				e,
				http.MethodPost,
				"/rooms/11111111-1111-1111-1111-111111111111/schedule/create",
				`{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`,
			)

			AssertJSONResponse(t, rec, tt.wantStatus, `{
				"error": {
					"code": "`+tt.wantCode+`",
					"message": "`+tt.wantMessage+`"
				}
			}`)
		})
	}
}
