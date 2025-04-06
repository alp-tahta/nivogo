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

// ReleaseInventoryHandler handles release inventory requests
type ReleaseInventoryHandler struct {
	BaseHandler
	reader *kafka.Reader
}

// NewReleaseInventoryHandler creates a new release inventory handler
func NewReleaseInventoryHandler(l *slog.Logger, r repository.RepositoryI, brokers string) (*ReleaseInventoryHandler, error) {
	// Create a reader for release inventory requests
	reader := kafka.NewReader(kafka.ReaderConfig{
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

	return &ReleaseInventoryHandler{
		BaseHandler: BaseHandler{
			l:      l,
			r:      r,
			writer: writer,
		},
		reader: reader,
	}, nil
}

// Start starts the release inventory handler
func (h *ReleaseInventoryHandler) Start(ctx context.Context) error {
	h.l.Info("Starting release inventory handler")

	go h.handleMessages(ctx)

	return nil
}

// Stop stops the release inventory handler
func (h *ReleaseInventoryHandler) Stop() error {
	if err := h.reader.Close(); err != nil {
		return fmt.Errorf("failed to close release reader: %w", err)
	}
	if err := h.writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// handleMessages handles incoming release inventory messages
func (h *ReleaseInventoryHandler) handleMessages(ctx context.Context) {
	h.l.Info("Starting to handle release inventory requests")

	for {
		select {
		case <-ctx.Done():
			h.l.Info("Context canceled, stopping release inventory handler")
			return
		default:
			// Read message
			h.l.Info("Waiting for release inventory request...")
			msg, err := h.reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					h.l.Info("Context canceled while reading message")
					return
				}
				h.l.Error("failed to read message", "error", err)
				continue
			}

			h.l.Info("Received release inventory request", "key", string(msg.Key), "value_length", len(msg.Value))

			// Parse request
			var request model.ReleaseInventoryRequest
			productID := request.ProductID
			orderID := request.OrderID
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				h.l.Error("failed to unmarshal request", "error", err)
				h.sendResponse(orderID, productID, false, "failed to unmarshal request")
				continue
			}

			h.l.Info("Processing release inventory request", "product_id", productID, "quantity", request.Quantity)

			// Process request using repository directly
			quantity, err := h.r.GetQuantityOfAProduct(productID)
			if err != nil {
				h.l.Error("failed to get quantity of product", "error", err, "product_id", productID)
				h.sendResponse(orderID, productID, false, fmt.Sprintf("failed to get quantity of product: %v", err))
				continue
			}

			// Release the quantity by adding it back
			quantity.Quantity = quantity.Quantity + request.Quantity

			err = h.r.UpdateQuantityOfAProduct(productID, quantity.Quantity)
			if err != nil {
				h.l.Error("failed to release inventory", "error", err, "product_id", productID)
				h.sendResponse(orderID, productID, false, fmt.Sprintf("failed to release inventory: %v", err))
				continue
			}

			// Send success response
			h.l.Info("Sending success response for release inventory", "product_id", productID)
			h.sendResponse(orderID, productID, true, "")
		}
	}
}
