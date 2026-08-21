package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/odeyaio/booking-service/internal/model"
)

const (
	queryBookingCreate = `
	INSERT INTO booking (id, slot_id, user_id, status_id)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at`

	queryBookingByID = `
	SELECT id, slot_id, user_id, status_id, created_at
	FROM booking
	WHERE id = $1`

	queryBookingList = `
	SELECT id, slot_id, user_id, status_id, created_at
	FROM booking
	ORDER BY created_at DESC, id
	LIMIT $1 OFFSET $2`

	queryBookingCount = `SELECT count(*) FROM booking`

	queryBookingListByUser = `
	SELECT b.id, b.slot_id, b.user_id, b.status_id, b.created_at
	FROM booking b
	JOIN slot s ON s.id = b.slot_id
	WHERE b.user_id = $1
	  AND s.start >= $2
	ORDER BY s.start, b.id`

	queryBookingCancel = `
	UPDATE booking
	SET status_id = $2
	WHERE id = $1
	RETURNING id, slot_id, user_id, status_id, created_at`
)

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	const op = "BookingRepository.Create"

	err := r.db.QueryRow(
		ctx,
		queryBookingCreate,
		booking.ID,
		booking.SlotID,
		booking.UserID,
		booking.StatusID,
	).Scan(&booking.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch {
			case pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "uq_active_booking_slot":
				err = ErrAlreadyExists
			case pgErr.Code == pgerrcode.ForeignKeyViolation && pgErr.ConstraintName == "booking_slot_id_fkey":
				err = ErrNotFound
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Booking, error) {
	const op = "BookingRepository.GetByID"

	rows, err := r.db.Query(ctx, queryBookingByID, id)
	if err != nil {
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	booking, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Booking])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}

func (r *BookingRepository) List(ctx context.Context, limit, offset int) ([]model.Booking, int, error) {
	const op = "BookingRepository.List"

	var total int
	if err := r.db.QueryRow(ctx, queryBookingCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%s count: %w", op, err)
	}

	rows, err := r.db.Query(ctx, queryBookingList, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Booking])
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, total, nil
}

func (r *BookingRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	from time.Time,
) ([]model.Booking, error) {
	const op = "BookingRepository.ListByUser"

	rows, err := r.db.Query(ctx, queryBookingListByUser, userID, from)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Booking])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return bookings, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, id uuid.UUID) (model.Booking, error) {
	const op = "BookingRepository.Cancel"

	rows, err := r.db.Query(ctx, queryBookingCancel, id, model.BookingStatusCancelled)
	if err != nil {
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	booking, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Booking])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrNotFound
		}
		return model.Booking{}, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}
