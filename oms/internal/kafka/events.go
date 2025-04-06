package kafka

import "time"

// InventoryEvent represents an inventory-related event
type InventoryEvent struct {
	OrderID   int       `json:"order_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
}
