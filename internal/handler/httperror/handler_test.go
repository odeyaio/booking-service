package httperror_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/handler/handlertest"
	"github.com/odeyaio/booking-service/internal/handler/httperror"
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
			err:        httperror.New(http.StatusBadRequest, httperror.CodeInvalidRequest, "invalid request"),
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
			e := echo.New()
			e.HTTPErrorHandler = httperror.NewHandler(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
			)
			e.GET("/test", func(c *echo.Context) error {
				return tt.err
			})

			rec := handlertest.PerformJSONRequest(
				t,
				e,
				http.MethodGet,
				"/test",
				"",
			)

			handlertest.AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}
