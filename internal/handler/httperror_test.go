package handler_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/handler"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "400 application error",
			err:        handler.NewHTTPError(http.StatusBadRequest, handler.ErrorCodeInvalidRequest, "invalid request"),
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"error": {
					"code": "INVALID_REQUEST",
					"message": "invalid request"
				}
			}`,
		},
		{
			name:       "500 unexpected error",
			err:        errors.New("database connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody: `{
				"error": {
					"code": "INTERNAL_ERROR",
					"message": "internal server error"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := handler.NewEcho()
			e.GET("/test", func(c *echo.Context) error {
				return tt.err
			})

			rec := handler.PerformJSONRequest(
				t,
				e,
				http.MethodGet,
				"/test",
				"",
			)

			handler.AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}
