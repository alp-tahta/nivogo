package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"oms/internal/handler"
	"oms/internal/logger"
	"oms/internal/repository"
	"oms/internal/routes"
	"oms/internal/server"
	"oms/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "postgres-oms"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "example"
	dbName     = "oms"
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
	svc, err := service.New(logger, repository)
	if err != nil {
		logger.Error("Failed to create service", "error", err)
		os.Exit(1)
	}
	defer svc.Close()

	handler := handler.New(logger, svc)

	routes.RegisterRoutes(mux, handler)

	// Create a channel to receive the server instance from server.Init
	serverReady := make(chan struct{})

	// Start the server in a goroutine
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

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
