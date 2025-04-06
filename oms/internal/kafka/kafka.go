package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"oms/internal/model"
	"oms/internal/repository"

	"github.com/segmentio/kafka-go"
)

const (
	// Topic names
	ReserveInventoryTopic  = "oms.reserve-inventory.0"
	ReleaseInventoryTopic  = "oms.release-inventory.0"
	InventoryResponseTopic = "oms.order-item-stock-reserved.0"
)

// KafkaClient handles Kafka operations
type KafkaClient struct {
	l        *slog.Logger
	r        repository.RepositoryI
	producer *InventoryProducer
	reader   *kafka.Reader
}

// New creates a new Kafka client
func New(l *slog.Logger, r repository.RepositoryI) (*KafkaClient, error) {
	// Get Kafka brokers from environment or use default
	brokers := []string{"localhost:9092"}
	if envBrokers := os.Getenv("KAFKA_BROKERS"); envBrokers != "" {
		brokers = []string{envBrokers}
	}

	// Create the inventory producer
	producer := NewInventoryProducer(l, brokers)

	// Create a reader for receiving responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    InventoryResponseTopic,
		GroupID:  "oms-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})

	return &KafkaClient{
		l:        l,
		r:        r,
		producer: producer,
		reader:   reader,
	}, nil
}

// Close closes the Kafka client
func (k *KafkaClient) Close() error {
	// Close the producer
	if err := k.producer.Close(); err != nil {
		k.l.Error("failed to close producer", "error", err)
		return fmt.Errorf("failed to close producer: %w", err)
	}

	// Close the reader
	if err := k.reader.Close(); err != nil {
		k.l.Error("failed to close reader", "error", err)
		return fmt.Errorf("failed to close reader: %w", err)
	}

	return nil
}

// ReserveInventory sends a reserve inventory request to Kafka
func (k *KafkaClient) ReserveInventory(orderID int, productID int, quantity int) error {
	return k.producer.ReserveInventory(orderID, productID, quantity)
}

// ReleaseInventory sends a release inventory request to Kafka
func (k *KafkaClient) ReleaseInventory(orderID int, productID int, quantity int) error {
	return k.producer.ReleaseInventory(orderID, productID, quantity)
}

// WaitForInventoryResponse waits for a response from the inventory service
func (k *KafkaClient) WaitForInventoryResponse(orderID int, productID int, timeout time.Duration) error {
	k.l.Info("Waiting for inventory response", "order_id", orderID, "product_id", productID, "timeout", timeout)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get broker addresses from the existing reader
	brokers := k.reader.Config().Brokers

	// Create a dedicated reader for this response
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  fmt.Sprintf("oms-response-%d-%d", orderID, productID), // Unique group ID for this request
		Topic:    InventoryResponseTopic,
		MaxBytes: 10e6, // 10MB
		MinBytes: 10e3, // 10KB
		MaxWait:  time.Second,
	})
	defer reader.Close()

	k.l.Info("Created dedicated reader for inventory response", "order_id", orderID, "product_id", productID, "brokers", brokers)

	// Read messages until we find a match or context is canceled
	for {
		select {
		case <-ctx.Done():
			k.l.Error("Context deadline exceeded while waiting for inventory response", "order_id", orderID, "product_id", productID)
			return fmt.Errorf("timeout waiting for inventory response: %w", ctx.Err())
		default:
			// Read message with a shorter timeout to allow for more frequent checking of context
			readCtx, readCancel := context.WithTimeout(ctx, 1*time.Second)
			msg, err := reader.ReadMessage(readCtx)
			readCancel()

			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					// This is just the short read timeout or context canceled, continue waiting
					continue
				}
				k.l.Error("failed to read message", "error", err)
				continue
			}

			k.l.Info("Received message from Kafka", "key", string(msg.Key), "value_length", len(msg.Value))

			// Check if this is the response we're looking for
			expectedKey := fmt.Sprintf("%d-%d", orderID, productID)
			if string(msg.Key) == expectedKey {
				k.l.Info("Found matching response", "order_id", orderID, "product_id", productID)
				var response model.InventoryResponse
				if err := json.Unmarshal(msg.Value, &response); err != nil {
					k.l.Error("Failed to unmarshal response", "error", err)
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}

				if !response.Success {
					k.l.Error("Inventory operation failed", "order_id", orderID, "product_id", productID, "error", response.Error)
					return fmt.Errorf("inventory operation failed: %s", response.Error)
				}

				k.l.Info("Received successful inventory response", "order_id", orderID, "product_id", productID)
				return nil
			} else {
				k.l.Info("Received response for different key", "expected", expectedKey, "received", string(msg.Key))
			}
		}
	}
}
