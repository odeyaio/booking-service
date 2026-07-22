package httperror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

const (
	CodeInvalidRequest = "INVALID_REQUEST"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeInternal       = "INTERNAL_ERROR"
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

func New(status int, code, message string) error {
	return &httpError{
		status:  status,
		code:    code,
		message: message,
	}
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

func NewHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		status := http.StatusInternalServerError
		code := CodeInternal
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
