package repository

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
)

const (
	queryRoomCreate = `
	INSERT INTO room (id, name, description, capacity)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at`

	queryRoomByID = `
	SELECT id, name, description, capacity, created_at
	FROM room
	WHERE id = $1`

	queryRoomList = `
	SELECT id, name, description, capacity, created_at
	FROM room
	ORDER BY created_at, id`
)

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	const op = "RoomRepository.Create"

	err := r.db.QueryRow(ctx, queryRoomCreate,
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

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Room, error) {
	const op = "RoomRepository.GetByID"

	rows, err := r.db.Query(ctx, queryRoomByID, id)
	if err != nil {
		return model.Room{}, fmt.Errorf("%s: %w", op, err)
	}

	room, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Room])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}

		return model.Room{}, fmt.Errorf("%s: %w", op, err)
	}

	return room, nil
}

func (r *RoomRepository) List(ctx context.Context) ([]model.Room, error) {
	const op = "RoomRepository.List"

	rows, err := r.db.Query(ctx, queryRoomList)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rooms, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Room])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
