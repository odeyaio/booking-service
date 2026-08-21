package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

type UserService struct {
	repo userRepository
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(
	ctx context.Context,
	email string,
	password string,
	roleID model.RoleID,
) (model.User, error) {
	const op = "UserService.Register"

	email, valid := normalizeEmail(email)
	if !valid || password == "" || !validRoleID(roleID) {
		return model.User{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return model.User{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
		}
		return model.User{}, fmt.Errorf("%s: hash password: %w", op, err)
	}

	user := model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		RoleID:       roleID,
	}
	if err := s.repo.Create(ctx, &user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			err = ErrEmailAlreadyExists
		}
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (model.User, error) {
	const op = "UserService.Login"

	email, valid := normalizeEmail(email)
	if !valid || password == "" {
		return model.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrInvalidCredentials
		}
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	return user, nil
}

func normalizeEmail(value string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(value))
	address, err := mail.ParseAddress(email)
	return email, err == nil && address.Address == email
}

func validRoleID(roleID model.RoleID) bool {
	return roleID == model.RoleAdmin || roleID == model.RoleUser
}
