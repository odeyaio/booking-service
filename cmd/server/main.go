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
	"github.com/odeyaio/booking-service/internal/handler/httperror"
	roomhandler "github.com/odeyaio/booking-service/internal/handler/room"
	"github.com/odeyaio/booking-service/internal/logger"
	roomrepository "github.com/odeyaio/booking-service/internal/repository/room"
	roomservice "github.com/odeyaio/booking-service/internal/service/room"
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

	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Error("failed to connect to db", "err", err)
	}

	e := echo.New()

	e.Logger = log
	e.HTTPErrorHandler = httperror.NewHandler(log)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/_info", func(c *echo.Context) error {
		return nil
	})

	roomRepo := roomrepository.New(db)
	roomService := roomservice.New(roomRepo)
	roomHandler := roomhandler.New(roomService)
	roomHandler.RegisterRoutes(e)

	server := &http.Server{
		Addr:         cfg.Addr,
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
