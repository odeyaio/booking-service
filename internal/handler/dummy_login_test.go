package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummyLoginHandler_Login(t *testing.T) {
	const secret = "test-secret"
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantUserID uuid.UUID
		wantRole   Role
	}{
		{
			name:       "admin token",
			body:       `{"role":"admin"}`,
			wantStatus: http.StatusOK,
			wantUserID: adminID,
			wantRole:   RoleAdmin,
		},
		{
			name:       "user token",
			body:       `{"role":"user"}`,
			wantStatus: http.StatusOK,
			wantUserID: userID,
			wantRole:   RoleUser,
		},
		{
			name:       "invalid role",
			body:       `{"role":"guest"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing role",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEcho()
			h := NewDummyLoginHandler(secret, adminID, userID)
			e.POST("/dummyLogin", h.Login)

			rec := PerformJSONRequest(t, e, http.MethodPost, "/dummyLogin", tt.body)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus != http.StatusOK {
				assert.JSONEq(t, `{"error":{"code":"INVALID_REQUEST","message":"invalid request"}}`, rec.Body.String())
				return
			}

			var response tokenResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			claims := new(AuthClaims)
			token, err := jwt.ParseWithClaims(response.Token, claims, func(*jwt.Token) (any, error) {
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			require.NoError(t, err)
			require.True(t, token.Valid)
			assert.Equal(t, tt.wantUserID, claims.UserID)
			assert.Equal(t, tt.wantRole, claims.Role)
		})
	}
}
