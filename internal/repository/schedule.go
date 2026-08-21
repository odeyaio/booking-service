package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

	queryScheduleList = `
	SELECT id, room_id, days_of_week, start_time, end_time
	FROM schedule
	ORDER BY room_id`
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

	schedule, err := pgx.CollectExactlyOneRow(rows, rowToSchedule)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Schedule{}, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		return model.Schedule{}, fmt.Errorf("%s: %w", op, err)
	}

	return schedule, nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]model.Schedule, error) {
	const op = "ScheduleRepository.List"

	rows, err := r.txGetter.DefaultTrOrDB(ctx, r.db).Query(ctx, queryScheduleList)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	schedules, err := pgx.CollectRows(rows, rowToSchedule)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return schedules, nil
}

func rowToSchedule(row pgx.CollectableRow) (model.Schedule, error) {
	var (
		schedule   model.Schedule
		days       []int16
		start, end pgtype.Time
	)

	if err := row.Scan(
		&schedule.ID,
		&schedule.RoomID,
		&days,
		&start,
		&end,
	); err != nil {
		return model.Schedule{}, err
	}

	schedule.DaysOfWeek = make([]model.Weekday, 0, len(days))
	for _, day := range days {
		weekday, err := model.NewWeekday(int(day))
		if err != nil {
			return model.Schedule{}, err
		}
		schedule.DaysOfWeek = append(schedule.DaysOfWeek, weekday)
	}

	var err error
	schedule.StartTime, err = timeOfDayFromPG(start)
	if err != nil {
		return model.Schedule{}, err
	}
	schedule.EndTime, err = timeOfDayFromPG(end)
	if err != nil {
		return model.Schedule{}, err
	}

	return schedule, nil
}

func timeOfDayFromPG(value pgtype.Time) (model.TimeOfDay, error) {
	if !value.Valid {
		return 0, errors.New("invalid null time")
	}

	minutes := value.Microseconds / int64(time.Minute/time.Microsecond)
	return model.NewTimeOfDay(int(minutes/60), int(minutes%60))
}
