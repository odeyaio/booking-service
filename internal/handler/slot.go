package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
)

type slotListAvailableRequest struct {
	RoomID uuid.UUID `param:"roomId"`
	Date   string    `query:"date"`
}

type slotResponse struct {
	ID     uuid.UUID `json:"id"`
	RoomID uuid.UUID `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type slotListAvailableResponse struct {
	Slots []slotResponse `json:"slots"`
}

func toSlotResponse(slot model.Slot) slotResponse {
	return slotResponse{
		ID:     slot.ID,
		RoomID: slot.RoomID,
		Start:  slot.Start,
		End:    slot.End,
	}
}

type slotService interface {
	ListAvailable(ctx context.Context, roomID uuid.UUID, dateString string) ([]model.Slot, error)
}

type SlotHandler struct {
	svc slotService
}

func NewSlotHandler(svc slotService) *SlotHandler {
	return &SlotHandler{svc: svc}
}

func (h *SlotHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/rooms/:roomId/slots/list", h.ListAvailable)
}

func (h *SlotHandler) ListAvailable(c *echo.Context) error {
	var req slotListAvailableRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	slots, err := h.svc.ListAvailable(c.Request().Context(), req.RoomID, req.Date)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
		}
	}

	slotResponses := make([]slotResponse, 0, len(slots))
	for _, slot := range slots {
		slotResponses = append(slotResponses, toSlotResponse(slot))
	}

	return c.JSON(http.StatusOK, slotListAvailableResponse{
		Slots: slotResponses,
	})
}
