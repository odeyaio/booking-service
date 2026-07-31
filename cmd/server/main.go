package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	roomHandler.RegisterRoutes(e)

	slotRepo := repository.NewSlotRepository(db)
	slotGenerator := service.NewSlotGenerator(slotRepo)
	slotService := service.NewSlotService(slotRepo)
	slotHandler := handler.NewSlotHandler(slotService)
	slotHandler.RegisterRoutes(e)

	scheduleRepo := repository.NewScheduleRepository(db)
	scheduleService := service.NewScheduleService(scheduleRepo, slotGenerator)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	scheduleHandler.RegisterRoutes(e)

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
