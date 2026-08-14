package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type dummyLoginRequest struct {
	Role Role `json:"role"`
}

type DummyLoginHandler struct {
	secret      []byte
	adminUserID uuid.UUID
	userUserID  uuid.UUID
}

func NewDummyLoginHandler(secret string, adminUserID, userUserID uuid.UUID) *DummyLoginHandler {
	return &DummyLoginHandler{
		secret:      []byte(secret),
		adminUserID: adminUserID,
		userUserID:  userUserID,
	}
}

func (h *DummyLoginHandler) Login(c *echo.Context) error {
	var req dummyLoginRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	var userID uuid.UUID
	switch req.Role {
	case RoleAdmin:
		userID = h.adminUserID
	case RoleUser:
		userID = h.userUserID
	default:
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	token, err := signToken(h.secret, userID, req.Role)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: token})
}
