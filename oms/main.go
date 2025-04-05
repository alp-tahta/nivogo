package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"oms/internal/logger"
	"oms/internal/repository"
	"oms/internal/server"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "postgres-product"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "example"
	dbName     = "product"
)

func main() {
	port := os.Getenv("PORT")
	// Initialize logger
	l := logger.Init()

	// Connect to database
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

	// Initialize repository
	repo := repository.New(l, db)

	// Create server
	srv := server.New(port, repo)

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			l.Error("server error", "error", err)
		}
	}()

	// Listen for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown
	stop()
	l.Info("shutting down gracefully, press Ctrl+C again to force")

	// Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("server forced to shutdown", "error", err)
	}

	l.Info("server exited properly")
}
