package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/odeyaio/booking-service/internal/config"
	"github.com/odeyaio/booking-service/internal/handler"
	"github.com/odeyaio/booking-service/internal/logger"
	"github.com/odeyaio/booking-service/internal/repository"
	"github.com/odeyaio/booking-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.Setup(cfg.Env)
	if err != nil {
		panic(err)
	}

	db, err := pgxpool.New(context.Background(), repository.ConnectionURL(cfg))
	if err != nil {
		log.Error("failed to connect to db", "err", err)
	}
	txManager := manager.Must(trmpgx.NewDefaultFactory(db))

	e := echo.New()

	e.Logger = log
	e.HTTPErrorHandler = handler.NewHTTPErrorHandler(log)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/_info", func(c *echo.Context) error {
		return nil
	})

	roomRepo := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	slotRepo := repository.NewSlotRepository(db)
	slotGenerator := service.NewSlotGenerator(slotRepo)
	slotService := service.NewSlotService(slotRepo, roomRepo)
	slotHandler := handler.NewSlotHandler(slotService)

	scheduleRepo := repository.NewScheduleRepository(db)
	scheduleService := service.NewScheduleService(scheduleRepo, slotGenerator, txManager)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	bookingRepo := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepo, slotRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)

	if cfg.Auth.DummyLogin.Enabled {
		authHandler := handler.NewAuthHandler(
			cfg.Auth.JWTSecret,
			cfg.Auth.DummyLogin.AdminUserID,
			cfg.Auth.DummyLogin.UserUserID,
		)
		e.POST("/dummyLogin", authHandler.DummyLogin)
	}

	authenticated := e.Group("", handler.JWTMiddleware(cfg.Auth.JWTSecret))
	authenticated.GET("/rooms/list", roomHandler.List)
	authenticated.GET("/rooms/:roomId/slots/list", slotHandler.ListAvailable)

	admin := authenticated.Group("", handler.RequireRole(handler.RoleAdmin))
	admin.POST("/rooms/create", roomHandler.Create)
	admin.POST("/rooms/:roomId/schedule/create", scheduleHandler.Create)
	admin.GET("/bookings/list", bookingHandler.List)

	user := authenticated.Group("", handler.RequireRole(handler.RoleUser))
	user.POST("/bookings/create", bookingHandler.Create)
	user.GET("/bookings/my", bookingHandler.ListMy)
	user.POST("/bookings/:bookingId/cancel", bookingHandler.Cancel)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      e,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		log.Info("starting server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "err", err)
	}

	log.Info("server stopped")
}
