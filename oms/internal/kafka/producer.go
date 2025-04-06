package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// InventoryProducer handles sending inventory-related events to Kafka
type InventoryProducer struct {
	l             *slog.Logger
	reserveWriter *kafka.Writer
	releaseWriter *kafka.Writer
}

// NewInventoryProducer creates a new inventory producer
func NewInventoryProducer(l *slog.Logger, brokers []string) *InventoryProducer {
	// Create writers for sending messages
	reserveWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        ReserveInventoryTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Ensure messages are written before returning
	}

	releaseWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        ReleaseInventoryTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Ensure messages are written before returning
	}

	return &InventoryProducer{
		l:             l,
		reserveWriter: reserveWriter,
		releaseWriter: releaseWriter,
	}
}

// ReserveInventory sends a reserve inventory request to Kafka
func (p *InventoryProducer) ReserveInventory(orderID int, productID int, quantity int) error {
	event := InventoryEvent{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal reserve inventory event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.reserveWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d-%d", orderID, productID)),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write reserve inventory message: %w", err)
	}

	p.l.Info("sent reserve inventory request", "order_id", orderID, "product_id", productID, "quantity", quantity)
	return nil
}

// ReleaseInventory sends a release inventory request to Kafka
func (p *InventoryProducer) ReleaseInventory(orderID int, productID int, quantity int) error {
	event := InventoryEvent{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal release inventory event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.releaseWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d-%d", orderID, productID)),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write release inventory message: %w", err)
	}

	p.l.Info("sent release inventory request", "order_id", orderID, "product_id", productID, "quantity", quantity)
	return nil
}

// Close closes the inventory producer
func (p *InventoryProducer) Close() error {
	if err := p.reserveWriter.Close(); err != nil {
		return fmt.Errorf("failed to close reserve writer: %w", err)
	}
	if err := p.releaseWriter.Close(); err != nil {
		return fmt.Errorf("failed to close release writer: %w", err)
	}
	return nil
}
