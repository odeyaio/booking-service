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

type roomCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type roomResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Capacity    *int      `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
}

type roomCreateResponse struct {
	Room roomResponse `json:"room"`
}

type roomListResponse struct {
	Rooms []roomResponse `json:"rooms"`
}

func toRoomResponse(room model.Room) roomResponse {
	return roomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
		CreatedAt:   room.CreatedAt,
	}
}

type roomService interface {
	Create(ctx context.Context, name string, description *string, capacity *int) (model.Room, error)
	List(ctx context.Context) ([]model.Room, error)
}

type RoomHandler struct {
	svc roomService
}

func NewRoomHandler(svc roomService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

func (h *RoomHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/rooms/list", h.List)
	e.POST("/rooms/create", h.Create)
}

func (h *RoomHandler) Create(c *echo.Context) error {
	var req roomCreateRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(
			http.StatusBadRequest,
			ErrorCodeInvalidRequest,
			"invalid request",
		)
	}

	room, err := h.svc.Create(c.Request().Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
		}

		return err
	}

	return c.JSON(http.StatusCreated, roomCreateResponse{
		Room: toRoomResponse(room),
	})
}

func (h *RoomHandler) List(c *echo.Context) error {
	rooms, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}

	resp := make([]roomResponse, 0, len(rooms))
	for _, room := range rooms {
		resp = append(resp, toRoomResponse(room))
	}

	return c.JSON(http.StatusOK, roomListResponse{
		Rooms: resp,
	})
}
