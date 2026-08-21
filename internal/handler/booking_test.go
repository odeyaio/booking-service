package handler

import (
	"errors"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/pagination"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func withPrincipal(principal Principal) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(principalContextKey, principal)
			return next(c)
		}
	}
}

func TestBookingHandler_Create(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	slotID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bookingID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	svc := NewMockBookingService(t)
	svc.EXPECT().
		Create(mock.Anything, userID, slotID).
		Return(model.Booking{
			ID:        bookingID,
			SlotID:    slotID,
			UserID:    userID,
			StatusID:  model.BookingStatusActive,
			CreatedAt: createdAt,
		}, nil).
		Once()
	e := NewEcho()
	e.POST("/bookings/create", NewBookingHandler(svc).Create, withPrincipal(Principal{
		UserID: userID,
		Role:   RoleUser,
	}))

	rec := PerformJSONRequest(t, e, http.MethodPost, "/bookings/create", `{
		"slotId": "11111111-1111-1111-1111-111111111111",
		"createConferenceLink": true
	}`)

	AssertJSONResponse(t, rec, http.StatusCreated, `{
		"booking": {
			"id": "22222222-2222-2222-2222-222222222222",
			"slotId": "11111111-1111-1111-1111-111111111111",
			"userId": "00000000-0000-0000-0000-000000000002",
			"status": "active",
			"createdAt": "2026-08-15T12:00:00Z"
		}
	}`)
}

func TestBookingHandler_List(t *testing.T) {
	bookingID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	svc := NewMockBookingService(t)
	svc.EXPECT().
		List(mock.Anything, pagination.Params{Page: 2, PageSize: 10}).
		Return([]model.Booking{{
			ID:       bookingID,
			StatusID: model.BookingStatusCancelled,
		}}, 11, nil).
		Once()
	e := NewEcho()
	e.GET("/bookings/list", NewBookingHandler(svc).List)

	rec := PerformJSONRequest(t, e, http.MethodGet, "/bookings/list?page=2&pageSize=10", "")

	AssertJSONResponse(t, rec, http.StatusOK, `{
		"bookings": [{
			"id": "22222222-2222-2222-2222-222222222222",
			"slotId": "00000000-0000-0000-0000-000000000000",
			"userId": "00000000-0000-0000-0000-000000000000",
			"status": "cancelled",
			"createdAt": "0001-01-01T00:00:00Z"
		}],
		"pagination": {"page": 2, "pageSize": 10, "total": 11}
	}`)
}

func TestBookingHandler_ListDefaults(t *testing.T) {
	svc := NewMockBookingService(t)
	svc.EXPECT().
		List(mock.Anything, pagination.Params{Page: pagination.DefaultPage, PageSize: pagination.DefaultPageSize}).
		Return(nil, 0, nil).
		Once()
	e := NewEcho()
	e.GET("/bookings/list", NewBookingHandler(svc).List)

	rec := PerformJSONRequest(t, e, http.MethodGet, "/bookings/list", "")

	AssertJSONResponse(t, rec, http.StatusOK, `{
		"bookings": [],
		"pagination": {"page": 1, "pageSize": 20, "total": 0}
	}`)
}

func TestBookingHandler_ListMyUsesPrincipal(t *testing.T) {
	userID := uuid.New()
	svc := NewMockBookingService(t)
	svc.EXPECT().
		ListMy(mock.Anything, userID).
		Return(nil, nil).
		Once()
	e := NewEcho()
	e.GET("/bookings/my", NewBookingHandler(svc).ListMy, withPrincipal(Principal{UserID: userID, Role: RoleUser}))

	rec := PerformJSONRequest(t, e, http.MethodGet, "/bookings/my", "")

	AssertJSONResponse(t, rec, http.StatusOK, `{"bookings":[]}`)
}

func TestBookingHandler_CancelForbidden(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	svc := NewMockBookingService(t)
	svc.EXPECT().
		Cancel(mock.Anything, userID, bookingID).
		Return(model.Booking{}, service.ErrForbidden).
		Once()
	e := NewEcho()
	e.POST("/bookings/:bookingId/cancel", NewBookingHandler(svc).Cancel, withPrincipal(Principal{
		UserID: userID,
		Role:   RoleUser,
	}))

	rec := PerformJSONRequest(
		t,
		e,
		http.MethodPost,
		"/bookings/22222222-2222-2222-2222-222222222222/cancel",
		"",
	)

	AssertJSONResponse(t, rec, http.StatusForbidden, `{
		"error": {
			"code": "FORBIDDEN",
			"message": "cannot cancel another user's booking"
		}
	}`)
}

func TestBookingHTTPError(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{service.ErrInvalidInput, http.StatusBadRequest, ErrorCodeInvalidRequest},
		{service.ErrSlotNotFound, http.StatusNotFound, ErrorCodeSlotNotFound},
		{service.ErrSlotAlreadyBooked, http.StatusConflict, ErrorCodeSlotAlreadyBooked},
		{service.ErrBookingNotFound, http.StatusNotFound, ErrorCodeBookingNotFound},
	}

	for _, tt := range tests {
		var appErr *httpError
		assert.True(t, errors.As(bookingHTTPError(tt.err), &appErr))
		assert.Equal(t, tt.wantStatus, appErr.status)
		assert.Equal(t, tt.wantCode, appErr.code)
	}
}
