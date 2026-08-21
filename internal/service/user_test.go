package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		email     string
		password  string
		roleID    model.RoleID
		setupMock func(*MockUserRepository)
		wantErr   error
	}{
		{
			name:     "user registered",
			email:    "  User@Example.com ",
			password: "secret-password",
			roleID:   model.RoleUser,
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(user *model.User) bool {
						return user.ID != uuid.Nil() &&
							user.Email == "user@example.com" &&
							user.RoleID == model.RoleUser &&
							bcrypt.CompareHashAndPassword(
								[]byte(user.PasswordHash),
								[]byte("secret-password"),
							) == nil
					})).
					Run(func(_ context.Context, user *model.User) {
						user.CreatedAt = createdAt
					}).
					Return(nil).
					Once()
			},
		},
		{
			name:     "email already registered",
			email:    "user@example.com",
			password: "secret-password",
			roleID:   model.RoleUser,
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.User")).
					Return(repository.ErrAlreadyExists).
					Once()
			},
			wantErr: ErrEmailAlreadyExists,
		},
		{
			name:     "invalid email",
			email:    "not-an-email",
			password: "secret-password",
			roleID:   model.RoleUser,
			wantErr:  ErrInvalidInput,
		},
		{
			name:    "empty password",
			email:   "user@example.com",
			roleID:  model.RoleUser,
			wantErr: ErrInvalidInput,
		},
		{
			name:     "invalid role",
			email:    "user@example.com",
			password: "secret-password",
			wantErr:  ErrInvalidInput,
		},
		{
			name:     "password is too long",
			email:    "user@example.com",
			password: strings.Repeat("a", 73),
			roleID:   model.RoleUser,
			wantErr:  ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepository(t)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			user, err := NewUserService(repo).Register(
				context.Background(),
				tt.email,
				tt.password,
				tt.roleID,
			)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, model.User{}, user)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "user@example.com", user.Email)
			assert.Equal(t, createdAt, user.CreatedAt)
			assert.NotEmpty(t, user.PasswordHash)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret-password"), bcrypt.MinCost)
	require.NoError(t, err)
	repoErr := errors.New("database unavailable")
	wantUser := model.User{
		ID:           uuid.NewV7(),
		Email:        "user@example.com",
		PasswordHash: string(passwordHash),
		RoleID:       model.RoleUser,
	}

	tests := []struct {
		name      string
		email     string
		password  string
		setupMock func(*MockUserRepository)
		wantErr   error
	}{
		{
			name:     "user authenticated",
			email:    " USER@example.com ",
			password: "secret-password",
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					GetByEmail(mock.Anything, "user@example.com").
					Return(wantUser, nil).
					Once()
			},
		},
		{
			name:     "user not found",
			email:    "missing@example.com",
			password: "secret-password",
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					GetByEmail(mock.Anything, "missing@example.com").
					Return(model.User{}, repository.ErrNotFound).
					Once()
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					GetByEmail(mock.Anything, "user@example.com").
					Return(wantUser, nil).
					Once()
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:     "invalid email",
			email:    "invalid",
			password: "secret-password",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name:     "repository failed",
			email:    "user@example.com",
			password: "secret-password",
			setupMock: func(repo *MockUserRepository) {
				repo.EXPECT().
					GetByEmail(mock.Anything, "user@example.com").
					Return(model.User{}, repoErr).
					Once()
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockUserRepository(t)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			user, err := NewUserService(repo).Login(context.Background(), tt.email, tt.password)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, model.User{}, user)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantUser, user)
		})
	}
}
