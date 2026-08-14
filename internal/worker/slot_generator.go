package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/odeyaio/booking-service/internal/model"
)

type scheduleRepository interface {
	List(ctx context.Context) ([]model.Schedule, error)
}

type slotGenerator interface {
	Generate(ctx context.Context, schedule model.Schedule, from, to time.Time) error
}

type SlotGeneratorWorker struct {
	scheduleRepo  scheduleRepository
	slotGenerator slotGenerator
	log           *slog.Logger
	interval      time.Duration
	window        time.Duration
	now           func() time.Time
}

func NewSlotGeneratorWorker(
	scheduleRepo scheduleRepository,
	slotGenerator slotGenerator,
	log *slog.Logger,
	interval time.Duration,
	window time.Duration,
) *SlotGeneratorWorker {
	return &SlotGeneratorWorker{
		scheduleRepo:  scheduleRepo,
		slotGenerator: slotGenerator,
		log:           log,
		interval:      interval,
		window:        window,
		now:           time.Now,
	}
}

func (w *SlotGeneratorWorker) Run(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	w.runOnceAndLog(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runOnceAndLog(ctx)
		}
	}
}

func (w *SlotGeneratorWorker) RunOnce(ctx context.Context) error {
	schedules, err := w.scheduleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list schedules: %w", err)
	}

	from := w.now().UTC()
	to := from.Add(w.window)
	errs := make([]error, 0)

	for _, schedule := range schedules {
		if err := w.slotGenerator.Generate(ctx, schedule, from, to); err != nil {
			errs = append(errs, fmt.Errorf("schedule %s: %w", schedule.ID, err))
		}
	}

	return errors.Join(errs...)
}

func (w *SlotGeneratorWorker) runOnceAndLog(ctx context.Context) {
	startedAt := time.Now()
	if err := w.RunOnce(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			w.log.Error("slot generation failed", "err", err)
		}
		return
	}

	w.log.Info("slot generation completed", "duration", time.Since(startedAt))
}
