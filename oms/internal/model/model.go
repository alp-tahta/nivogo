package model

import "time"

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OrderItem struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

type Order struct {
	ID        int         `json:"id"`
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// CreateOrder represents an order without an ID, used for creating new orders
type CreateOrder struct {
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
}

type OrderSaga struct {
	OrderID   int
	Status    string
	Step      int
	CreatedAt time.Time
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Kafka-related models
type ReserveInventoryRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ReleaseInventoryRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type InventoryResponse struct {
	ProductID int    `json:"product_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}
