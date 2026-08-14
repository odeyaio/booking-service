package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	ErrorCodeInvalidRequest    = "INVALID_REQUEST"
	ErrorCodeUnauthorized      = "UNAUTHORIZED"
	ErrorCodeForbidden         = "FORBIDDEN"
	ErrorCodeNotFound          = "NOT_FOUND"
	ErrorCodeSlotNotFound      = "SLOT_NOT_FOUND"
	ErrorCodeSlotAlreadyBooked = "SLOT_ALREADY_BOOKED"
	ErrorCodeBookingNotFound   = "BOOKING_NOT_FOUND"
	ErrorCodeRoomNotFound      = "ROOM_NOT_FOUND"
	ErrorCodeScheduleExists    = "SCHEDULE_EXISTS"
	ErrorCodeInternal          = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}

type httpError struct {
	status  int
	code    string
	message string
}

func NewHTTPError(status int, code, message string) error {
	return &httpError{
		status:  status,
		code:    code,
		message: message,
	}
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

func NewHTTPErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		response, unwrapErr := echo.UnwrapResponse(c.Response())
		if unwrapErr == nil && response.Committed {
			return
		}

		status := http.StatusInternalServerError
		code := ErrorCodeInternal
		message := "internal server error"

		if appErr, ok := errors.AsType[*httpError](err); ok {
			status = appErr.status
			code = appErr.code
			message = appErr.message
		} else {
			log.Error("request failed", "err", err)
		}

		_ = c.JSON(status, NewErrorResponse(code, message))
	}
}
