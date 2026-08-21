package handler

import (
	"context"
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
)

type userService interface {
	Register(ctx context.Context, email, password string, roleID model.RoleID) (model.User, error)
	Login(ctx context.Context, email, password string) (model.User, error)
}

type UserHandler struct {
	svc    userService
	secret []byte
}

func NewUserHandler(svc userService, secret string) *UserHandler {
	return &UserHandler{svc: svc, secret: []byte(secret)}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type registerResponse struct {
	User userResponse `json:"user"`
}

func (h *UserHandler) Register(c *echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	roleID, ok := roleIDFromRole(req.Role)
	if !ok {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	user, err := h.svc.Register(c.Request().Context(), req.Email, req.Password, roleID)
	if err != nil {
		return userHTTPError(err)
	}

	return c.JSON(http.StatusCreated, registerResponse{User: toUserResponse(user)})
}

func (h *UserHandler) Login(c *echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "invalid credentials")
	}

	user, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return userHTTPError(err)
	}

	role, ok := roleFromRoleID(user.RoleID)
	if !ok {
		return errors.New("unknown user role")
	}
	token, err := signToken(h.secret, user.ID, role)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: token})
}

func toUserResponse(user model.User) userResponse {
	role, _ := roleFromRoleID(user.RoleID)
	return userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      role,
		CreatedAt: user.CreatedAt,
	}
}

func roleIDFromRole(role Role) (model.RoleID, bool) {
	switch role {
	case RoleAdmin:
		return model.RoleAdmin, true
	case RoleUser:
		return model.RoleUser, true
	default:
		return 0, false
	}
}

func roleFromRoleID(roleID model.RoleID) (Role, bool) {
	switch roleID {
	case model.RoleAdmin:
		return RoleAdmin, true
	case model.RoleUser:
		return RoleUser, true
	default:
		return "", false
	}
}

func userHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "email already registered")
	case errors.Is(err, service.ErrInvalidCredentials):
		return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "invalid credentials")
	default:
		return err
	}
}
