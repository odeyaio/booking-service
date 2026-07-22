package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/handler/httperror"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
)

type createRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

type response struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Capacity    *int      `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
}

type createResponse struct {
	Room response `json:"room"`
}

type listResponse struct {
	Rooms []response `json:"rooms"`
}

func toResponse(room model.Room) response {
	return response{
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

type Handler struct {
	svc roomService
}

func New(svc roomService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/rooms/list", h.List)
	e.POST("/rooms/create", h.Create)
}

func (h *Handler) Create(c *echo.Context) error {
	var req createRequest
	if err := c.Bind(&req); err != nil {
		return httperror.New(
			http.StatusBadRequest,
			httperror.CodeInvalidRequest,
			"invalid request",
		)
	}

	room, err := h.svc.Create(c.Request().Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return httperror.New(http.StatusBadRequest, httperror.CodeInvalidRequest, "invalid request")
		}

		return err
	}

	return c.JSON(http.StatusCreated, createResponse{
		Room: toResponse(room),
	})
}

func (h *Handler) List(c *echo.Context) error {
	rooms, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}

	resp := make([]response, 0, len(rooms))
	for _, room := range rooms {
		resp = append(resp, toResponse(room))
	}

	return c.JSON(http.StatusOK, listResponse{
		Rooms: resp,
	})
}
