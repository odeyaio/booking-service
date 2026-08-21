package service

import (
	"context"
	"fmt"
	"strings"
	"uuid"

	"github.com/odeyaio/booking-service/internal/model"
)

type roomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	List(ctx context.Context) ([]model.Room, error)
}

type RoomService struct {
	repo roomRepository
}

func NewRoomService(repo roomRepository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) Create(ctx context.Context, name string, description *string, capacity *int) (model.Room, error) {
	const op = "RoomService.Create"

	name = strings.TrimSpace(name)
	if name == "" {
		return model.Room{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
	}

	if capacity != nil && *capacity <= 0 {
		return model.Room{}, fmt.Errorf("%s: %w", op, ErrInvalidInput)
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

func (s *RoomService) List(ctx context.Context) ([]model.Room, error) {
	const op = "RoomService.List"

	rooms, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
