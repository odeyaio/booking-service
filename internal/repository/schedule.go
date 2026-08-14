package repository

import (
	"context"
	"errors"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
)

const (
	queryScheduleCreate = `
	INSERT INTO schedule (id, room_id, days_of_week, start_time, end_time)
	VALUES ($1, $2, $3, $4, $5)`

	queryScheduleByRoomID = `
	SELECT id, room_id, days_of_week, start_time, end_time
	FROM schedule
	WHERE room_id = $1`
)

type ScheduleRepository struct {
	db       *pgxpool.Pool
	txGetter *trmpgx.CtxGetter
}

func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{
		db:       db,
		txGetter: trmpgx.DefaultCtxGetter,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule model.Schedule) error {
	const op = "ScheduleRepository.Create"

	_, err := r.txGetter.DefaultTrOrDB(ctx, r.db).Exec(
		ctx,
		queryScheduleCreate,
		schedule.ID,
		schedule.RoomID,
		schedule.DaysOfWeek,
		schedule.StartTime,
		schedule.EndTime)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				err = ErrAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				err = ErrNotFound
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (model.Schedule, error) {
	const op = "ScheduleRepository.GetByRoomID"

	rows, err := r.txGetter.DefaultTrOrDB(ctx, r.db).Query(ctx, queryScheduleByRoomID, roomID)
	if err != nil {
		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	schedule, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Schedule])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Schedule{}, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	return schedule, nil
}
