package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"oms/internal/model"

	"github.com/segmentio/kafka-go"
)

const (
	// Topic names
	ReserveInventoryTopic  = "oms.reserve-inventory.0"
	ReleaseInventoryTopic  = "oms.release-inventory.0"
	InventoryResponseTopic = "oms.inventory-response.0"
)

// KafkaClient handles Kafka operations
type KafkaClient struct {
	l             *slog.Logger
	reserveWriter *kafka.Writer
	releaseWriter *kafka.Writer
	reader        *kafka.Reader
}

// New creates a new Kafka client
func New(l *slog.Logger) (*KafkaClient, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	// Create writers for sending messages
	reserveWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        ReserveInventoryTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Ensure messages are written before returning
	}

	releaseWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        ReleaseInventoryTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Ensure messages are written before returning
	}

	// Create a reader for receiving responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    InventoryResponseTopic,
		GroupID:  "oms-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})

	return &KafkaClient{
		l:             l,
		reserveWriter: reserveWriter,
		releaseWriter: releaseWriter,
		reader:        reader,
	}, nil
}

// Close closes the Kafka client
func (k *KafkaClient) Close() error {
	if err := k.reserveWriter.Close(); err != nil {
		return fmt.Errorf("failed to close reserve writer: %w", err)
	}
	if err := k.releaseWriter.Close(); err != nil {
		return fmt.Errorf("failed to close release writer: %w", err)
	}
	if err := k.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	return nil
}

// ReserveInventory sends a reserve inventory request to Kafka
func (k *KafkaClient) ReserveInventory(productID int, quantity int) error {
	// Create request
	request := model.ReserveInventoryRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a context with timeout for writing the message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send message
	err = k.reserveWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", productID)),
		Value: jsonData,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	k.l.Info("sent reserve inventory request", "product_id", productID, "quantity", quantity)
	return nil
}

// ReleaseInventory sends a release inventory request to Kafka
func (k *KafkaClient) ReleaseInventory(productID int, quantity int) error {
	// Create request
	request := model.ReleaseInventoryRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a context with timeout for writing the message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send message
	err = k.releaseWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", productID)),
		Value: jsonData,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	k.l.Info("sent release inventory request", "product_id", productID, "quantity", quantity)
	return nil
}

// WaitForInventoryResponse waits for a response from the inventory service
func (k *KafkaClient) WaitForInventoryResponse(productID int, timeout time.Duration) error {
	k.l.Info("Waiting for inventory response", "product_id", productID, "timeout", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a ticker to log waiting status
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Keep track of messages that don't match our product ID
	skippedMessages := 0
	maxSkippedMessages := 100 // Limit the number of messages we'll skip before timing out

	for {
		select {
		case <-ctx.Done():
			k.l.Error("Context deadline exceeded while waiting for inventory response", "product_id", productID)
			return fmt.Errorf("timeout waiting for inventory response: %w", ctx.Err())
		case <-ticker.C:
			k.l.Info("Still waiting for inventory response", "product_id", productID, "timeout", timeout, "skipped_messages", skippedMessages)
		default:
			// Read message with a shorter timeout to allow for more frequent checking of other cases
			readCtx, readCancel := context.WithTimeout(ctx, 1*time.Second)
			msg, err := k.reader.ReadMessage(readCtx)
			readCancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// This is just the short read timeout, continue waiting
					continue
				}
				if err == context.Canceled {
					// The main context was canceled
					k.l.Error("Context canceled while waiting for inventory response", "product_id", productID)
					return fmt.Errorf("timeout waiting for inventory response: %w", err)
				}
				k.l.Error("failed to read message", "error", err)
				continue
			}

			k.l.Info("Received message from Kafka", "key", string(msg.Key), "value_length", len(msg.Value))

			// Check if this is the response we're looking for
			if string(msg.Key) == fmt.Sprintf("%d", productID) {
				k.l.Info("Found matching response for product", "product_id", productID)
				var response model.InventoryResponse
				if err := json.Unmarshal(msg.Value, &response); err != nil {
					k.l.Error("Failed to unmarshal response", "error", err)
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}

				if !response.Success {
					k.l.Error("Inventory operation failed", "product_id", productID, "error", response.Error)
					return fmt.Errorf("inventory operation failed: %s", response.Error)
				}

				k.l.Info("Received successful inventory response", "product_id", productID)
				return nil
			} else {
				k.l.Info("Received response for different product", "expected", productID, "received", string(msg.Key))
				skippedMessages++
				if skippedMessages >= maxSkippedMessages {
					k.l.Error("Too many non-matching messages received", "product_id", productID, "skipped_messages", skippedMessages)
					return fmt.Errorf("timeout after skipping %d non-matching messages", skippedMessages)
				}
			}
		}
	}
}
