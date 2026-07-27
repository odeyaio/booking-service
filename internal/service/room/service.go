package room

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
)

type roomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	List(ctx context.Context) ([]model.Room, error)
}

type Service struct {
	repo roomRepository
}

func New(repo roomRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name string, description *string, capacity *int) (model.Room, error) {
	const op = "room.Service.Create"

	name = strings.TrimSpace(name)
	if name == "" {
		return model.Room{}, fmt.Errorf("%s: %w", op, service.ErrInvalidInput)
	}

	if capacity != nil && *capacity <= 0 {
		return model.Room{}, fmt.Errorf("%s: %w", op, service.ErrInvalidInput)
	}

	room := model.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}

	if err := s.repo.Create(ctx, &room); err != nil {
		return model.Room{}, fmt.Errorf("%s: %w", op, err)
	}

	return room, nil
}

func (s *Service) List(ctx context.Context) ([]model.Room, error) {
	const op = "room.Service.List"

	rooms, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
