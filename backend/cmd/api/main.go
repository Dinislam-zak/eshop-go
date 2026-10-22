package main

import (
	"context"
	"log"
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
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: failed to load .env file: %v (using environment variables or defaults)", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	txManager := repository.NewTransactionManager(db)

	productService := service.NewProductService(productRepo, txManager)
	orderService := service.NewOrderService(orderRepo, productRepo, txManager)

	router := httpTransport.NewRouter(cfg, productService, orderService)

	srv := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ln, err := net.Listen("tcp", cfg.GetServerAddress())
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", cfg.GetServerAddress(), err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	log.Printf("server started on %s", cfg.GetServerAddress())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	case <-quit:
		log.Println("shutting down server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("failed to close database connection: %v", err)
	}

	log.Println("server exiting")
}