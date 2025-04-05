package model

import "time"

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OrderItem struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

type Order struct {
	ID        string      `json:"id"`
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type OrderSaga struct {
	OrderID   string
	Status    string
	Step      int
	CreatedAt time.Time
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type ReserveInventoryRequest struct {
	Quantity int `json:"quantity"`
}

type ReleaseInventoryRequest struct {
	Quantity int `json:"quantity"`
}
