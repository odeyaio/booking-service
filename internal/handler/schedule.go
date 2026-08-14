package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
)

type scheduleCreateRequest struct {
	RoomID     uuid.UUID `param:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

type scheduleResponse struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

type scheduleCreateResponse struct {
	Schedule scheduleResponse `json:"schedule"`
}

func toScheduleResponse(schedule model.Schedule) scheduleResponse {
	daysOfWeek := make([]int, 0, len(schedule.DaysOfWeek))
	for _, day := range schedule.DaysOfWeek {
		daysOfWeek = append(daysOfWeek, int(day))
	}

	return scheduleResponse{
		ID:         schedule.ID,
		RoomID:     schedule.RoomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  schedule.StartTime.String(),
		EndTime:    schedule.EndTime.String(),
	}
}

type scheduleService interface {
	Create(
		ctx context.Context,
		roomID uuid.UUID,
		daysOfWeek []int,
		startTime string,
		endTime string) (model.Schedule, error)
}

type ScheduleHandler struct {
	svc scheduleService
}

func NewScheduleHandler(svc scheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

func (h *ScheduleHandler) Create(c *echo.Context) error {
	var req scheduleCreateRequest
	if err := c.Bind(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
	}

	schedule, err := h.svc.Create(
		c.Request().Context(),
		req.RoomID,
		req.DaysOfWeek,
		req.StartTime,
		req.EndTime)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			return NewHTTPError(http.StatusBadRequest, ErrorCodeInvalidRequest, "invalid request")
		}

		return err
	}

	return c.JSON(http.StatusCreated, scheduleCreateResponse{
		Schedule: toScheduleResponse(schedule),
	})
}
