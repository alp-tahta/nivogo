package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"product/internal/handler"
	"product/internal/logger"
	"product/internal/repository"
	"product/internal/routes"
	"product/internal/server"
	"product/internal/service"
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

	// Define the connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

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

	logger.Info("Starting HTTP server at", "port", port)
	err = server.Init(port, mux)
	if err != nil {
		logger.Error("Failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
