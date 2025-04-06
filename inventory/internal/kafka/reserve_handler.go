package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"inventory/internal/model"
	"inventory/internal/repository"

	"github.com/segmentio/kafka-go"
)

// ReserveInventoryHandler handles reserve inventory requests
type ReserveInventoryHandler struct {
	BaseHandler
	reader *kafka.Reader
}

// NewReserveInventoryHandler creates a new reserve inventory handler
func NewReserveInventoryHandler(l *slog.Logger, r repository.RepositoryI, brokers string) (*ReserveInventoryHandler, error) {
	// Create a reader for reserve inventory requests
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    ReserveInventoryTopic,
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

	return &ReserveInventoryHandler{
		BaseHandler: BaseHandler{
			l:      l,
			r:      r,
			writer: writer,
		},
		reader: reader,
	}, nil
}

// Start starts the reserve inventory handler
func (h *ReserveInventoryHandler) Start(ctx context.Context) error {
	h.l.Info("Starting reserve inventory handler")

	go h.handleMessages(ctx)

	return nil
}

// Stop stops the reserve inventory handler
func (h *ReserveInventoryHandler) Stop() error {
	if err := h.reader.Close(); err != nil {
		return fmt.Errorf("failed to close reserve reader: %w", err)
	}
	if err := h.writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// handleMessages handles incoming reserve inventory messages
func (h *ReserveInventoryHandler) handleMessages(ctx context.Context) {
	h.l.Info("Starting to handle reserve inventory requests")

	for {
		select {
		case <-ctx.Done():
			h.l.Info("Context canceled, stopping reserve inventory handler")
			return
		default:
			// Read message
			h.l.Info("Waiting for reserve inventory request...")
			msg, err := h.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					h.l.Info("Context canceled while reading message")
					return
				}
				h.l.Error("failed to read message", "error", err)
				continue
			}

			h.l.Info("Received reserve inventory request", "key", string(msg.Key), "value_length", len(msg.Value))

			// Parse product ID from key
			productID := 0
			fmt.Sscanf(string(msg.Key), "%d", &productID)

			// Parse request
			var request model.ReserveInventoryRequest
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				h.l.Error("failed to unmarshal request", "error", err)
				h.sendResponse(productID, false, "failed to unmarshal request")
				continue
			}

			h.l.Info("Processing reserve inventory request", "product_id", productID, "quantity", request.Quantity)

			// Process request using repository directly
			quantity, err := h.r.GetQuantityOfAProduct(productID)
			if err != nil {
				h.l.Error("failed to get quantity of product", "error", err, "product_id", productID)
				h.sendResponse(productID, false, fmt.Sprintf("failed to get quantity of product: %v", err))
				continue
			}

			if quantity.Quantity < request.Quantity {
				h.l.Error("not enough quantity available", "product_id", productID, "available", quantity.Quantity, "requested", request.Quantity)
				h.sendResponse(productID, false, "not enough quantity available")
				continue
			}

			// Reserve the quantity by reducing it
			quantity.Quantity = quantity.Quantity - request.Quantity

			err = h.r.UpdateQuantityOfAProduct(productID, quantity.Quantity)
			if err != nil {
				h.l.Error("failed to reserve inventory", "error", err, "product_id", productID)
				h.sendResponse(productID, false, fmt.Sprintf("failed to reserve inventory: %v", err))
				continue
			}

			// Send success response
			h.l.Info("Sending success response for reserve inventory", "product_id", productID)
			h.sendResponse(productID, true, "")
		}
	}
}
