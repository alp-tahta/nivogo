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

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	l      *slog.Logger
	r      repository.RepositoryI
	writer *kafka.Writer
}

// sendResponse sends a response to Kafka with retry logic
func (h *BaseHandler) sendResponse(orderID int, productID int, success bool, errorMsg string) {
	h.l.Info("Preparing to send inventory response", "order_id", orderID, "product_id", productID, "success", success)

	response := model.InventoryResponse{
		OrderID:   orderID,
		ProductID: productID,
		Success:   success,
		Error:     errorMsg,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.l.Error("failed to marshal response", "error", err)
		return
	}

	// Try to send the response with retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Create a context with timeout for writing the message
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		h.l.Info("Writing response to Kafka", "topic", InventoryResponseTopic, "key", fmt.Sprintf("%d", productID), "attempt", i+1)
		err = h.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("%d", productID)),
			Value: jsonData,
		})
		if err == nil {
			h.l.Info("Successfully sent inventory response", "product_id", productID, "success", success, "attempt", i+1)
			return
		}

		h.l.Error("failed to write response", "error", err, "attempt", i+1)
		if i < maxRetries-1 {
			// Wait before retrying, using exponential backoff
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}

	h.l.Error("failed to send response after all retries", "product_id", productID, "retries", maxRetries)
}
