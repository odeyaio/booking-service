package handler

import (
	"context"
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/pagination"
	"github.com/odeyaio/booking-service/internal/service"
)

const (
	bookingStatusActive    = "active"
	bookingStatusCancelled = "cancelled"
)

type bookingCreateRequest struct {
	SlotID uuid.UUID `json:"slotId"`
}

type bookingListRequest struct {
	Page     *int `query:"page"`
	PageSize *int `query:"pageSize"`
}

type bookingCancelRequest struct {
	BookingID uuid.UUID `param:"bookingId"`
}

type bookingResponse struct {
	ID        uuid.UUID `json:"id"`
	SlotID    uuid.UUID `json:"slotId"`
	UserID    uuid.UUID `json:"userId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type bookingCreateResponse struct {
	Booking bookingResponse `json:"booking"`
}

type bookingListResponse struct {
	Bookings   []bookingResponse   `json:"bookings"`
	Pagination pagination.Metadata `json:"pagination"`
}

type bookingMyResponse struct {
	Bookings []bookingResponse `json:"bookings"`
}

func toBookingResponse(booking model.Booking) bookingResponse {
	status := bookingStatusActive
	if booking.StatusID == model.BookingStatusCancelled {
		status = bookingStatusCancelled
	}

	return bookingResponse{
		ID:        booking.ID,
		SlotID:    booking.SlotID,
		UserID:    booking.UserID,
		Status:    status,
		CreatedAt: booking.CreatedAt,
	}
}

func toBookingResponses(bookings []model.Booking) []bookingResponse {
	response := make([]bookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		response = append(response, toBookingResponse(booking))
	}
	return response
}

type bookingService interface {
	Create(ctx context.Context, userID, slotID uuid.UUID) (model.Booking, error)
	List(ctx context.Context, params pagination.Params) ([]model.Booking, int, error)
	ListMy(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
	Cancel(ctx context.Context, userID, bookingID uuid.UUID) (model.Booking, error)
}

type BookingHandler struct {
	svc bookingService
}

func NewBookingHandler(svc bookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

func (h *BookingHandler) Create(c *echo.Context) error {
	var req bookingCreateRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	principal, ok := PrincipalFromContext(c)
	if !ok {
		return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
	}

	booking, err := h.svc.Create(c.Request().Context(), principal.UserID, req.SlotID)
	if err != nil {
		return bookingHTTPError(err)
	}

	return c.JSON(http.StatusCreated, bookingCreateResponse{
		Booking: toBookingResponse(booking),
	})
}

func (h *BookingHandler) List(c *echo.Context) error {
	var req bookingListRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	params := pagination.New(req.Page, req.PageSize)

	bookings, total, err := h.svc.List(c.Request().Context(), params)
	if err != nil {
		return bookingHTTPError(err)
	}

	return c.JSON(http.StatusOK, bookingListResponse{
		Bookings:   toBookingResponses(bookings),
		Pagination: pagination.NewMetadata(params, total),
	})
}

func (h *BookingHandler) ListMy(c *echo.Context) error {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
	}

	bookings, err := h.svc.ListMy(c.Request().Context(), principal.UserID)
	if err != nil {
		return bookingHTTPError(err)
	}

	return c.JSON(http.StatusOK, bookingMyResponse{
		Bookings: toBookingResponses(bookings),
	})
}

func (h *BookingHandler) Cancel(c *echo.Context) error {
	var req bookingCancelRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	principal, ok := PrincipalFromContext(c)
	if !ok {
		return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
	}

	booking, err := h.svc.Cancel(c.Request().Context(), principal.UserID, req.BookingID)
	if err != nil {
		return bookingHTTPError(err)
	}

	return c.JSON(http.StatusOK, bookingCreateResponse{
		Booking: toBookingResponse(booking),
	})
}

func bookingHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	case errors.Is(err, service.ErrSlotNotFound):
		return NewHTTPError(http.StatusNotFound, ErrorCodeSlotNotFound, "slot not found")
	case errors.Is(err, service.ErrSlotAlreadyBooked):
		return NewHTTPError(http.StatusConflict, ErrorCodeSlotAlreadyBooked, "slot is already booked")
	case errors.Is(err, service.ErrBookingNotFound):
		return NewHTTPError(http.StatusNotFound, ErrorCodeBookingNotFound, "booking not found")
	case errors.Is(err, service.ErrForbidden):
		return NewHTTPError(http.StatusForbidden, ErrorCodeForbidden, "cannot cancel another user's booking")
	default:
		return err
	}
}
