package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
)

const (
	queryUserCreate = `
	INSERT INTO "user" (id, email, password_hash, role_id)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at`

	queryUserByEmail = `
	SELECT id, email, password_hash, role_id, created_at
	FROM "user"
	WHERE email = $1`
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	const op = "UserRepository.Create"

	err := r.db.QueryRow(
		ctx,
		queryUserCreate,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.RoleID,
	).Scan(&user.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "user_email_key" {
			err = ErrAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	const op = "UserRepository.GetByEmail"

	rows, err := r.db.Query(ctx, queryUserByEmail, email)
	if err != nil {
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}
		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
