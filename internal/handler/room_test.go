package handler

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/odeyaio/booking-service/internal/model"
	"github.com/odeyaio/booking-service/internal/service"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*MockRoomService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "201 room created",
			body: `{"name":"Room 1"}`,
			setupMock: func(svc *MockRoomService) {
				svc.EXPECT().
					Create(
						mock.Anything,
						"Room 1",
						(*string)(nil),
						(*int)(nil),
					).
					Return(model.Room{
						ID:   uuid.MustParse("12341234-1234-1234-1234-123412341234"),
						Name: "Room 1",
					}, nil).
					Once()
			},
			wantStatus: http.StatusCreated,
			wantBody: `{
				"room": {
					"id": "12341234-1234-1234-1234-123412341234",
					"name": "Room 1",
					"description": null,
					"capacity": null,
					"createdAt": "0001-01-01T00:00:00Z"
				}
			}`,
		},
		{
			name: "400 invalid service input",
			body: `{"name":""}`,
			setupMock: func(svc *MockRoomService) {
				svc.EXPECT().
					Create(
						mock.Anything,
						"",
						(*string)(nil),
						(*int)(nil),
					).
					Return(model.Room{}, service.ErrInvalidInput).
					Once()
			},
			wantStatus: http.StatusBadRequest,
			wantBody: `{
				"error": {
					"code": "INVALID_REQUEST",
					"message": "invalid request"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockRoomService(t)
			tt.setupMock(svc)

			e := NewEcho()
			e.POST("/rooms/create", NewRoomHandler(svc).Create)

			rec := PerformJSONRequest(
				t,
				e,
				http.MethodPost,
				"/rooms/create",
				tt.body,
			)

			AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*MockRoomService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "200 rooms list",
			setupMock: func(svc *MockRoomService) {
				svc.EXPECT().
					List(mock.Anything).
					Return([]model.Room{{
						ID:   uuid.MustParse("12341234-1234-1234-1234-123412341234"),
						Name: "Room 1",
					}}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody: `{
				"rooms": [
					{
						"id": "12341234-1234-1234-1234-123412341234",
						"name": "Room 1",
						"description": null,
						"capacity": null,
						"createdAt": "0001-01-01T00:00:00Z"
					}
				]
			}`,
		},
		{
			name: "200 empty list",
			setupMock: func(svc *MockRoomService) {
				svc.EXPECT().
					List(mock.Anything).
					Return([]model.Room{}, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"rooms":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockRoomService(t)
			tt.setupMock(svc)

			e := NewEcho()
			e.GET("/rooms/list", NewRoomHandler(svc).List)

			rec := PerformJSONRequest(
				t,
				e,
				http.MethodGet,
				"/rooms/list",
				"",
			)

			AssertJSONResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}
