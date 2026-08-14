package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware(t *testing.T) {
	const secret = "test-secret"
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name          string
		withToken     bool
		claims        jwt.MapClaims
		wantStatus    int
		wantPrincipal *Principal
	}{
		{
			name:       "missing token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing user id",
			withToken:  true,
			claims:     jwt.MapClaims{"role": RoleUser},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "malformed user id",
			withToken: true,
			claims: jwt.MapClaims{
				"user_id": "not-a-uuid",
				"role":    RoleUser,
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing role",
			withToken:  true,
			claims:     jwt.MapClaims{"user_id": userID.String()},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "unknown role",
			withToken: true,
			claims: jwt.MapClaims{
				"user_id": userID.String(),
				"role":    "guest",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "valid claims",
			withToken: true,
			claims: jwt.MapClaims{
				"user_id": userID.String(),
				"role":    RoleUser,
			},
			wantStatus: http.StatusNoContent,
			wantPrincipal: &Principal{
				UserID: userID,
				Role:   RoleUser,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEcho()
			e.GET("/protected", func(c *echo.Context) error {
				principal, ok := PrincipalFromContext(c)
				require.True(t, ok)
				assert.Equal(t, *tt.wantPrincipal, principal)
				return c.NoContent(http.StatusNoContent)
			}, JWTMiddleware(secret))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.withToken {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)
				signedToken, err := token.SignedString([]byte(secret))
				require.NoError(t, err)
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+signedToken)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if tt.wantStatus == http.StatusNoContent {
				assert.Equal(t, tt.wantStatus, rec.Code)
				return
			}

			AssertJSONResponse(t, rec, tt.wantStatus, `{
					"error": {
						"code": "UNAUTHORIZED",
						"message": "unauthorized"
					}
				}`)
		})
	}
}

func TestRequireRole(t *testing.T) {
	const secret = "test-secret"
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name       string
		role       Role
		header     bool
		wantStatus int
		wantCode   string
	}{
		{
			name:       "admin is allowed",
			role:       RoleAdmin,
			header:     true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "user is forbidden",
			role:       RoleUser,
			header:     true,
			wantStatus: http.StatusForbidden,
			wantCode:   ErrorCodeForbidden,
		},
		{
			name:       "missing token is unauthorized",
			role:       RoleAdmin,
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrorCodeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEcho()
			e.GET("/admin", func(c *echo.Context) error {
				principal, ok := PrincipalFromContext(c)
				require.True(t, ok)
				assert.Equal(t, userID, principal.UserID)
				assert.Equal(t, tt.role, principal.Role)
				return c.NoContent(http.StatusNoContent)
			}, JWTMiddleware(secret), RequireRole(RoleAdmin))

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.header {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, AuthClaims{
					UserID: userID,
					Role:   tt.role,
				})
				signedToken, err := token.SignedString([]byte(secret))
				require.NoError(t, err)
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+signedToken)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantCode != "" {
				assert.JSONEq(t, `{
					"error": {
						"code": "`+tt.wantCode+`",
						"message": "`+strings.ToLower(http.StatusText(tt.wantStatus))+`"
					}
				}`, rec.Body.String())
			}
		})
	}
}
