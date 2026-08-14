package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_Register(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		body       string
		setupMock  func(*MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "user registered",
			body: `{"email":"user@example.com","password":"secret-password","role":"user"}`,
			setupMock: func(svc *MockUserService) {
				svc.EXPECT().
					Register(mock.Anything, "user@example.com", "secret-password", model.RoleUser).
					Return(model.User{
						ID:        userID,
						Email:     "user@example.com",
						RoleID:    model.RoleUser,
						CreatedAt: createdAt,
					}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody: `{
				"user": {
					"id":"11111111-1111-1111-1111-111111111111",
					"email":"user@example.com",
					"role":"user",
					"createdAt":"2026-08-15T12:00:00Z"
				}
			}`,
		},
		{
			name: "email already registered",
			body: `{"email":"user@example.com","password":"secret-password","role":"user"}`,
			setupMock: func(svc *MockUserService) {
				svc.EXPECT().
					Register(mock.Anything, "user@example.com", "secret-password", model.RoleUser).
					Return(model.User{}, service.ErrEmailAlreadyExists).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"INVALID_REQUEST","message":"email already registered"}}`,
		},
		{
			name:       "invalid role",
			body:       `{"email":"user@example.com","password":"secret-password","role":"guest"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"INVALID_REQUEST","message":"invalid request"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockUserService(t)
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}
			e := NewEcho()
			e.POST("/register", NewUserHandler(svc, "test-secret").Register)

			rec := PerformJSONRequest(t, e, http.MethodPost, "/register", tt.body)
			AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	const secret = "test-secret"
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name       string
		setupMock  func(*MockUserService)
		wantStatus int
	}{
		{
			name: "user authenticated",
			setupMock: func(svc *MockUserService) {
				svc.EXPECT().
					Login(mock.Anything, "user@example.com", "secret-password").
					Return(model.User{ID: userID, RoleID: model.RoleUser}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			setupMock: func(svc *MockUserService) {
				svc.EXPECT().
					Login(mock.Anything, "user@example.com", "secret-password").
					Return(model.User{}, service.ErrInvalidCredentials).
					Once()
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockUserService(t)
			tt.setupMock(svc)
			e := NewEcho()
			e.POST("/login", NewUserHandler(svc, secret).Login)

			rec := PerformJSONRequest(
				t,
				e,
				http.MethodPost,
				"/login",
				`{"email":"user@example.com","password":"secret-password"}`,
			)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusUnauthorized {
				assert.JSONEq(t, `{"error":{"code":"UNAUTHORIZED","message":"invalid credentials"}}`, rec.Body.String())
				return
			}

			var response tokenResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			claims := new(AuthClaims)
			token, err := jwt.ParseWithClaims(response.Token, claims, func(*jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			require.NoError(t, err)
			require.True(t, token.Valid)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, RoleUser, claims.Role)
		})
	}
}
