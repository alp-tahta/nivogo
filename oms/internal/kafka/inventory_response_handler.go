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

			// Parse product ID from key
			productID := 0
			fmt.Sscanf(string(msg.Key), "%d", &productID)

			// Parse response
			var response model.InventoryResponse
			if err := json.Unmarshal(msg.Value, &response); err != nil {
				h.l.Error("failed to unmarshal response", "error", err)
				continue
			}

			h.l.Info("Processing inventory response", "product_id", productID, "success", response.Success)

			// Process response
			if response.Success {
				h.l.Info("Inventory operation successful", "product_id", productID)

				// Find the saga for this product
				sagas, err := h.r.GetSagasByStatus("INVENTORY_RESERVED")
				if err != nil {
					h.l.Error("failed to get sagas", "error", err)
					continue
				}

				// Update saga status if found
				for _, saga := range sagas {
					// Check if this saga is waiting for this product
					order, err := h.r.GetOrder(saga.OrderID)
					if err != nil {
						h.l.Error("failed to get order", "error", err, "order_id", saga.OrderID)
						continue
					}

					// Check if any item in the order matches this product
					for _, item := range order.Items {
						if item.Product.ID == productID {
							// Update saga status
							err = h.r.UpdateSagaStatus(saga.OrderID, "COMPLETED")
							if err != nil {
								h.l.Error("failed to update saga status", "error", err, "order_id", saga.OrderID)
							} else {
								h.l.Info("Updated saga status to COMPLETED", "order_id", saga.OrderID)
							}
							break
						}
					}
				}
			} else {
				h.l.Error("Inventory operation failed", "product_id", productID, "error", response.Error)

				// Find the saga for this product
				sagas, err := h.r.GetSagasByStatus("INVENTORY_RESERVED")
				if err != nil {
					h.l.Error("failed to get sagas", "error", err)
					continue
				}

				// Update saga status if found
				for _, saga := range sagas {
					// Check if this saga is waiting for this product
					order, err := h.r.GetOrder(saga.OrderID)
					if err != nil {
						h.l.Error("failed to get order", "error", err, "order_id", saga.OrderID)
						continue
					}

					// Check if any item in the order matches this product
					for _, item := range order.Items {
						if item.Product.ID == productID {
							// Update saga status
							err = h.r.UpdateSagaStatus(saga.OrderID, "FAILED")
							if err != nil {
								h.l.Error("failed to update saga status", "error", err, "order_id", saga.OrderID)
							} else {
								h.l.Info("Updated saga status to FAILED", "order_id", saga.OrderID)
							}
							break
						}
					}
				}
			}
		}
	}
}
