package room

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/repository"
)

const (
	queryCreate = `
	INSERT INTO room (id, name, description, capacity)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at`

	queryByID = `
	SELECT id, name, description, capacity, created_at
	FROM room
	WHERE id = $1`

	queryList = `
	SELECT id, name, description, capacity, created_at
	FROM room
	ORDER BY created_at, id`
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, room *model.Room) error {
	const op = "RoomRepository.Create"

	err := r.db.QueryRow(ctx, queryCreate,
		room.ID,
		room.Name,
		room.Description,
		room.Capacity).
		Scan(&room.CreatedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (model.Room, error) {
	const op = "RoomRepository.GetByID"

	rows, err := r.db.Query(ctx, queryByID, id)
	if err != nil {
		return model.Room{}, fmt.Errorf("%s: %w", op, err)
	}

	room, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Room])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = repository.ErrNotFound
		}

		return model.Room{}, fmt.Errorf("%s: %w", op, err)
	}

	return room, nil
}

func (r *Repository) List(ctx context.Context) ([]model.Room, error) {
	const op = "RoomRepository.List"

	rows, err := r.db.Query(ctx, queryList)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rooms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Room])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
