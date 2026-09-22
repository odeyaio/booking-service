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

func TestHandlerEchoErrors(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "404 unknown route",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
			wantBody: `{
				"error": {
					"code": "NOT_FOUND",
					"message": "not found"
				}
			}`,
		},
		{
			name:       "405 wrong method",
			method:     http.MethodPost,
			path:       "/test",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody: `{
				"error": {
					"code": "INVALID_REQUEST",
					"message": "method not allowed"
				}
			}`,
		},
		{
			name:       "500 echo server error stays generic",
			method:     http.MethodGet,
			path:       "/server-error",
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
			e.GET("/test", func(_ *echo.Context) error {
				return nil
			})
			e.GET("/server-error", func(_ *echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "db is down")
			})

			rec := handler.PerformJSONRequest(t, e, tt.method, tt.path, "")

			handler.AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}
