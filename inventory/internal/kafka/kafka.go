package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"inventory/internal/repository"
)

// KafkaServer handles Kafka operations
type KafkaServer struct {
	l                       *slog.Logger
	r                       repository.RepositoryI
	reserveInventoryHandler *ReserveInventoryHandler
	releaseInventoryHandler *ReleaseInventoryHandler
}

// New creates a new Kafka server
func New(l *slog.Logger, r repository.RepositoryI) (*KafkaServer, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	// Create handlers
	reserveHandler, err := NewReserveInventoryHandler(l, r, brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create reserve inventory handler: %w", err)
	}

	releaseHandler, err := NewReleaseInventoryHandler(l, r, brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to create release inventory handler: %w", err)
	}

	return &KafkaServer{
		l:                       l,
		r:                       r,
		reserveInventoryHandler: reserveHandler,
		releaseInventoryHandler: releaseHandler,
	}, nil
}

// Close closes the Kafka server
func (k *KafkaServer) Close() error {
	if err := k.reserveInventoryHandler.Stop(); err != nil {
		return fmt.Errorf("failed to stop reserve inventory handler: %w", err)
	}
	if err := k.releaseInventoryHandler.Stop(); err != nil {
		return fmt.Errorf("failed to stop release inventory handler: %w", err)
	}
	return nil
}

// Start starts the Kafka server
func (k *KafkaServer) Start(ctx context.Context) error {
	k.l.Info("Starting Kafka server")

	// Start handlers
	if err := k.reserveInventoryHandler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start reserve inventory handler: %w", err)
	}

	if err := k.releaseInventoryHandler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start release inventory handler: %w", err)
	}

	return nil
}
