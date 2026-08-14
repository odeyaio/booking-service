package handler

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"

	principalContextKey = "principal"
)

type dummyLoginRequest struct {
	Role Role `json:"role"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   Role      `json:"role"`
	jwt.RegisteredClaims
}

func (c AuthClaims) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}

	switch c.Role {
	case RoleAdmin, RoleUser:
		return nil
	default:
		return errors.New("invalid role")
	}
}

type Principal struct {
	UserID uuid.UUID
	Role   Role
}

func PrincipalFromContext(c *echo.Context) (Principal, bool) {
	principal, ok := c.Get(principalContextKey).(Principal)
	return principal, ok
}

type AuthHandler struct {
	secret      []byte
	adminUserID uuid.UUID
	userUserID  uuid.UUID
}

func NewAuthHandler(secret string, adminUserID, userUserID uuid.UUID) *AuthHandler {
	return &AuthHandler{
		secret:      []byte(secret),
		adminUserID: adminUserID,
		userUserID:  userUserID,
	}
}

func (h *AuthHandler) DummyLogin(c *echo.Context) error {
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AuthClaims{
		UserID: userID,
		Role:   req.Role,
	})
	signedToken, err := token.SignedString(h.secret)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: signedToken})
}

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(secret),
		NewClaimsFunc: func(_ *echo.Context) jwt.Claims {
			return new(AuthClaims)
		},
		SuccessHandler: func(c *echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
			}

			claims, ok := token.Claims.(*AuthClaims)
			if !ok {
				return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
			}

			c.Set(principalContextKey, Principal{
				UserID: claims.UserID,
				Role:   claims.Role,
			})
			return nil
		},
		ErrorHandler: func(_ *echo.Context, _ error) error {
			return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
		},
	})
}

func RequireRole(role Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			principal, ok := PrincipalFromContext(c)
			if !ok {
				return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, "unauthorized")
			}

			if principal.Role != role {
				return NewHTTPError(http.StatusForbidden, ErrorCodeForbidden, "forbidden")
			}

			return next(c)
		}
	}
}
