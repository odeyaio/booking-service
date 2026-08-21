package service

import (
	"context"
	"fmt"
	"slices"
	"time"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
)

const slotDuration = 30 * time.Minute

type slotCreator interface {
	UpsertBatch(ctx context.Context, slots []model.Slot) error
}

type SlotGenerator struct {
	slotCreator slotCreator
}

func NewSlotGenerator(slotCreator slotCreator) *SlotGenerator {
	return &SlotGenerator{
		slotCreator: slotCreator,
	}
}

func (g *SlotGenerator) Generate(
	ctx context.Context,
	schedule model.Schedule,
	from, to time.Time,
) error {
	const op = "SlotGenerator.Generate"

	from = from.UTC()
	to = to.UTC()

	if !from.Before(to) {
		return fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	scheduleDuration := time.Duration(
		int(schedule.EndTime)-int(schedule.StartTime),
	) * time.Minute

	slotCountPerDay := int(scheduleDuration / slotDuration)
	dayCount := int(to.Sub(startOfDay(from))/(24*time.Hour)) + 1

	slots := make([]model.Slot, 0, slotCountPerDay*dayCount)

	for date := startOfDay(from); date.Before(to); date = date.AddDate(0, 0, 1) {
		weekday, err := model.WeekdayFromTime(date.Weekday())
		if err != nil {
			return fmt.Errorf("%s: %w", op, ErrInvalidInput)
		}

		if !slices.Contains(schedule.DaysOfWeek, weekday) {
			continue
		}

		dayStart := schedule.StartTime.At(date)
		dayEnd := schedule.EndTime.At(date)

		for slotStart := dayStart; !slotStart.Add(slotDuration).After(dayEnd); slotStart = slotStart.Add(slotDuration) {
			slotEnd := slotStart.Add(slotDuration)

			if slotStart.Before(from) || slotEnd.After(to) {
				continue
			}

			slots = append(slots, model.Slot{
				ID:         uuid.NewV7(),
				RoomID:     schedule.RoomID,
				ScheduleID: schedule.ID,
				Start:      slotStart,
				End:        slotEnd,
			})
		}
	}

	if len(slots) == 0 {
		return nil
	}

	if err := g.slotCreator.UpsertBatch(ctx, slots); err != nil {
		return fmt.Errorf("%s: create slots: %w", op, err)
	}

	return nil
}

func startOfDay(t time.Time) time.Time {
	t = t.UTC()

	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
