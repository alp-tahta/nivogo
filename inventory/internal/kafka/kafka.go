package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"inventory/internal/model"
	"inventory/internal/service"

	"github.com/segmentio/kafka-go"
)

const (
	// Topic names
	ReserveInventoryTopic  = "oms.reserve-inventory.0"
	ReleaseInventoryTopic  = "oms.release-inventory.0"
	InventoryResponseTopic = "oms.inventory-response.0"
)

// KafkaServer handles Kafka operations
type KafkaServer struct {
	l             *slog.Logger
	s             service.ServiceI
	reserveReader *kafka.Reader
	releaseReader *kafka.Reader
	writer        *kafka.Writer
}

// New creates a new Kafka server
func New(l *slog.Logger, s service.ServiceI) (*KafkaServer, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	// Create a reader for reserve inventory requests
	reserveReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    ReserveInventoryTopic,
		GroupID:  "inventory-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})

	// Create a reader for release inventory requests
	releaseReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    ReleaseInventoryTopic,
		GroupID:  "inventory-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})

	// Create a writer for responses
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        InventoryResponseTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}

	return &KafkaServer{
		l:             l,
		s:             s,
		reserveReader: reserveReader,
		releaseReader: releaseReader,
		writer:        writer,
	}, nil
}

// Close closes the Kafka server
func (k *KafkaServer) Close() error {
	if err := k.reserveReader.Close(); err != nil {
		return fmt.Errorf("failed to close reserve reader: %w", err)
	}
	if err := k.releaseReader.Close(); err != nil {
		return fmt.Errorf("failed to close release reader: %w", err)
	}
	if err := k.writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// Start starts the Kafka server
func (k *KafkaServer) Start(ctx context.Context) error {
	k.l.Info("Starting Kafka server")

	// Start a goroutine to handle reserve inventory requests
	go k.handleReserveInventory(ctx)

	// Start a goroutine to handle release inventory requests
	go k.handleReleaseInventory(ctx)

	return nil
}

// handleReserveInventory handles reserve inventory requests
func (k *KafkaServer) handleReserveInventory(ctx context.Context) {
	k.l.Info("Starting to handle reserve inventory requests")

	for {
		select {
		case <-ctx.Done():
			k.l.Info("Context canceled, stopping reserve inventory handler")
			return
		default:
			// Read message
			k.l.Info("Waiting for reserve inventory request...")
			msg, err := k.reserveReader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					k.l.Info("Context canceled while reading message")
					return
				}
				k.l.Error("failed to read message", "error", err)
				continue
			}

			k.l.Info("Received reserve inventory request", "key", string(msg.Key), "value_length", len(msg.Value))

			// Parse product ID from key
			productID := 0
			fmt.Sscanf(string(msg.Key), "%d", &productID)

			// Parse request
			var request model.ReserveInventoryRequest
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				k.l.Error("failed to unmarshal request", "error", err)
				k.sendResponse(productID, false, "failed to unmarshal request")
				continue
			}

			k.l.Info("Processing reserve inventory request", "product_id", productID, "quantity", request.Quantity)

			// Process request
			err = k.s.ReserveInventory(productID, request)
			if err != nil {
				k.l.Error("failed to reserve inventory", "error", err, "product_id", productID)
				k.sendResponse(productID, false, err.Error())
				continue
			}

			// Send success response
			k.l.Info("Sending success response for reserve inventory", "product_id", productID)
			k.sendResponse(productID, true, "")
		}
	}
}

// handleReleaseInventory handles release inventory requests
func (k *KafkaServer) handleReleaseInventory(ctx context.Context) {
	k.l.Info("Starting to handle release inventory requests")

	for {
		select {
		case <-ctx.Done():
			k.l.Info("Context canceled, stopping release inventory handler")
			return
		default:
			// Read message
			k.l.Info("Waiting for release inventory request...")
			msg, err := k.releaseReader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					k.l.Info("Context canceled while reading message")
					return
				}
				k.l.Error("failed to read message", "error", err)
				continue
			}

			k.l.Info("Received release inventory request", "key", string(msg.Key), "value_length", len(msg.Value))

			// Parse product ID from key
			productID := 0
			fmt.Sscanf(string(msg.Key), "%d", &productID)

			// Parse request
			var request model.ReleaseInventoryRequest
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				k.l.Error("failed to unmarshal request", "error", err)
				k.sendResponse(productID, false, "failed to unmarshal request")
				continue
			}

			k.l.Info("Processing release inventory request", "product_id", productID, "quantity", request.Quantity)

			// Process request
			err = k.s.ReleaseInventory(productID, request)
			if err != nil {
				k.l.Error("failed to release inventory", "error", err, "product_id", productID)
				k.sendResponse(productID, false, err.Error())
				continue
			}

			// Send success response
			k.l.Info("Sending success response for release inventory", "product_id", productID)
			k.sendResponse(productID, true, "")
		}
	}
}

// sendResponse sends a response to Kafka
func (k *KafkaServer) sendResponse(productID int, success bool, errorMsg string) {
	k.l.Info("Preparing to send inventory response", "product_id", productID, "success", success)

	response := model.InventoryResponse{
		ProductID: productID,
		Success:   success,
		Error:     errorMsg,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		k.l.Error("failed to marshal response", "error", err)
		return
	}

	// Try to send the response with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Create a context with timeout for writing the message
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		k.l.Info("Writing response to Kafka", "topic", InventoryResponseTopic, "key", fmt.Sprintf("%d", productID), "attempt", i+1)
		err = k.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("%d", productID)),
			Value: jsonData,
		})
		if err == nil {
			k.l.Info("Successfully sent inventory response", "product_id", productID, "success", success, "attempt", i+1)
			return
		}

		k.l.Error("failed to write response", "error", err, "attempt", i+1)
		if i < maxRetries-1 {
			// Wait before retrying, using exponential backoff
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}

	k.l.Error("failed to send response after all retries", "product_id", productID, "retries", maxRetries)
}
