package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/shop/backend/internal/config"
	"github.com/shop/backend/internal/database"
	"github.com/shop/backend/internal/repository"
	"github.com/shop/backend/internal/service"
	httpTransport "github.com/shop/backend/internal/transport/http"
)

func main() {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("failed to load .env file", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level.Set(cfg.GetLogLevel())

	db, err := database.NewDB(cfg)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get underlying sql.DB", "error", err)
		os.Exit(1)
	}

	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	txManager := repository.NewTransactionManager(db)

	productService := service.NewProductService(productRepo, txManager)
	orderService := service.NewOrderService(orderRepo, productRepo, txManager)

	router := httpTransport.NewRouter(cfg, sqlDB, productService, orderService)

	srv := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ln, err := net.Listen("tcp", cfg.GetServerAddress())
	if err != nil {
		slog.Error("failed to listen", "address", cfg.GetServerAddress(), "error", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	slog.Info("server started", "address", cfg.GetServerAddress())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-quit:
		slog.Info("shutting down server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	if err := sqlDB.Close(); err != nil {
		slog.Error("failed to close database connection", "error", err)
	}

	slog.Info("server exiting")
}