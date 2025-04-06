package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"oms/internal/model"
	"oms/internal/repository"
	"time"

	"github.com/segmentio/kafka-go"
)

// InventoryResponseHandler handles processing inventory responses
type InventoryResponseHandler struct {
	l      *slog.Logger
	r      repository.RepositoryI
	reader *kafka.Reader
}

// NewInventoryResponseHandler creates a new inventory response handler
func NewInventoryResponseHandler(l *slog.Logger, r repository.RepositoryI, brokers []string) *InventoryResponseHandler {
	// Create a reader for receiving responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    InventoryResponseTopic,
		GroupID:  "oms-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Second,
	})

	return &InventoryResponseHandler{
		l:      l,
		r:      r,
		reader: reader,
	}
}

// Close closes the inventory response handler
func (h *InventoryResponseHandler) Close() error {
	if err := h.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	return nil
}

// Start starts the inventory response handler
func (h *InventoryResponseHandler) Start(ctx context.Context) error {
	h.l.Info("Starting inventory response handler")

	go h.handleMessages(ctx)

	return nil
}

// handleMessages handles incoming inventory response messages
func (h *InventoryResponseHandler) handleMessages(ctx context.Context) {
	h.l.Info("Starting to handle inventory responses")

	for {
		select {
		case <-ctx.Done():
			h.l.Info("Context canceled, stopping inventory response handler")
			return
		default:
			// Read message
			h.l.Info("Waiting for inventory response...")
			msg, err := h.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					h.l.Info("Context canceled while reading message")
					return
				}
				h.l.Error("failed to read message", "error", err)
				continue
			}

			h.l.Info("Received inventory response", "key", string(msg.Key), "value_length", len(msg.Value))

			// Parse response
			var response model.InventoryResponse
			if err := json.Unmarshal(msg.Value, &response); err != nil {
				h.l.Error("failed to unmarshal response", "error", err)
				continue
			}

			h.l.Info("Processing inventory response", "order_id", response.OrderID, "product_id", response.ProductID, "success", response.Success)

			// Process response
			if response.Success {
				h.l.Info("Inventory operation successful", "order_id", response.OrderID, "product_id", response.ProductID)

				// Update saga status
				err = h.r.UpdateSagaStatus(response.OrderID, "COMPLETED")
				if err != nil {
					h.l.Error("failed to update saga status", "error", err, "order_id", response.OrderID)
				} else {
					h.l.Info("Updated saga status to COMPLETED", "order_id", response.OrderID)
				}
			} else {
				h.l.Error("Inventory operation failed", "order_id", response.OrderID, "product_id", response.ProductID, "error", response.Error)

				// Update saga status
				err = h.r.UpdateSagaStatus(response.OrderID, "FAILED")
				if err != nil {
					h.l.Error("failed to update saga status", "error", err, "order_id", response.OrderID)
				} else {
					h.l.Info("Updated saga status to FAILED", "order_id", response.OrderID)
				}
			}
		}
	}
}
