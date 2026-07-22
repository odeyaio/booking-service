package handlertest

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/handler/httperror"
	"github.com/stretchr/testify/assert"
)

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = httperror.NewHandler(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	return e
}

func PerformJSONRequest(
	t *testing.T,
	e *echo.Echo,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

func AssertJSONResponse(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	wantStatus int,
	wantBody string,
) {
	t.Helper()

	assert.Equal(t, wantStatus, rec.Code)
	assert.JSONEq(t, wantBody, rec.Body.String())
}
