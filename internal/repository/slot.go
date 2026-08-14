package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
)

const querySlotListAvailable = `
	SELECT s.id, s.room_id, s.schedule_id, s.start, s."end"
	FROM slot s
	WHERE s.room_id = $1
	  AND s.start >= $2
	  AND s.start < $3
	  AND NOT EXISTS (
		SELECT 1
		FROM booking b
		WHERE b.slot_id = s.id
		  AND b.status_id = 1
	  )
	ORDER BY s.start`

const querySlotByID = `
	SELECT id, room_id, schedule_id, start, "end"
	FROM slot
	WHERE id = $1`

type SlotRepository struct {
	db       *pgxpool.Pool
	txGetter *trmpgx.CtxGetter
}

func NewSlotRepository(db *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{
		db:       db,
		txGetter: trmpgx.DefaultCtxGetter,
	}
}

func (r *SlotRepository) CreateBatch(ctx context.Context, slots []model.Slot) error {
	const op = "SlotRepository.CreateBatch"

	rows := make([][]any, 0, len(slots))
	for _, slot := range slots {
		rows = append(rows, []any{
			slot.ID,
			slot.RoomID,
			slot.ScheduleID,
			slot.Start,
			slot.End,
		})
	}

	if len(rows) == 0 {
		return nil
	}

	_, err := r.txGetter.DefaultTrOrDB(ctx, r.db).CopyFrom(
		ctx,
		pgx.Identifier{"slot"},
		[]string{"id", "room_id", "schedule_id", "start", "end"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Slot, error) {
	const op = "SlotRepository.GetByID"

	rows, err := r.txGetter.DefaultTrOrDB(ctx, r.db).Query(ctx, querySlotByID, id)
	if err != nil {
		return model.Slot{}, fmt.Errorf("%s: %w", op, err)
	}

	slot, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Slot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}
		return model.Slot{}, fmt.Errorf("%s: %w", op, err)
	}

	return slot, nil
}

func (r *SlotRepository) ListAvailable(
	ctx context.Context,
	roomID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]model.Slot, error) {
	const op = "SlotRepository.ListAvailable"

	rows, err := r.txGetter.DefaultTrOrDB(ctx, r.db).Query(ctx, querySlotListAvailable, roomID, from, to)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slots, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Slot])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return slots, nil
}
