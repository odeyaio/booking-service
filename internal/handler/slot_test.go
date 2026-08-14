package handler

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/mock"
)

func TestSlotHandler_RoomNotFound(t *testing.T) {
	roomID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc := NewMockSlotService(t)
	svc.EXPECT().
		ListAvailable(mock.Anything, roomID, "2026-08-15").
		Return([]model.Slot(nil), service.ErrRoomNotFound).
		Once()
	e := NewEcho()
	e.GET("/rooms/:roomId/slots/list", NewSlotHandler(svc).ListAvailable)

	rec := PerformJSONRequest(
		t,
		e,
		http.MethodGet,
		"/rooms/11111111-1111-1111-1111-111111111111/slots/list?date=2026-08-15",
		"",
	)

	AssertJSONResponse(t, rec, http.StatusNotFound, `{
		"error": {
			"code": "ROOM_NOT_FOUND",
			"message": "room not found"
		}
	}`)
}
