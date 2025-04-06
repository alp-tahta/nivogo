package main

import (
	"context"
	"database/sql"
	"fmt"
	"inventory/internal/handler"
	"inventory/internal/kafka"
	"inventory/internal/logger"
	"inventory/internal/repository"
	"inventory/internal/routes"
	"inventory/internal/server"
	"inventory/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "postgres-inventory"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "example"
	dbName     = "inventory"
)

func main() {
	port := os.Getenv("PORT")

	// Define the connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	log.Println(connStr)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Ensure the connection is available
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger := logger.Init()

	mux := http.NewServeMux()

	repository := repository.New(logger, db)
	service := service.New(logger, repository)
	handler := handler.New(logger, service)

	routes.RegisterRoutes(mux, handler)

	// Create Kafka server
	kafkaServer, err := kafka.New(logger, service)
	if err != nil {
		logger.Error("Failed to create Kafka server", "error", err)
		os.Exit(1)
	}
	defer kafkaServer.Close()

	// Create a context for the Kafka server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the Kafka server
	if err := kafkaServer.Start(ctx); err != nil {
		logger.Error("Failed to start Kafka server", "error", err)
		os.Exit(1)
	}

	// Create a channel to receive the server instance from server.Init
	serverReady := make(chan struct{})

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info("Starting HTTP server at", "port", port)
		err := server.Init(port, mux)
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
		close(serverReady)
	}()

	// Set up signal notification channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either server to be ready or a signal
	select {
	case <-serverReady:
		logger.Info("Server started successfully")
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", "signal", sig)
		os.Exit(0)
	}

	// Wait for interrupt signal
	sig := <-sigChan
	logger.Info("Received signal, shutting down", "signal", sig)

	// Cancel the context to stop the Kafka server
	cancel()

	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
