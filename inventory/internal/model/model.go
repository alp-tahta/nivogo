package model

type CreateInventoryRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type QuantityOfAProduct struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Inventory struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ReduceQuantityOfAProduct struct {
	Amount int `json:"amount"`
}

type ReserveInventoryRequest struct {
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ReleaseInventoryRequest struct {
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Kafka-related models
type InventoryResponse struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}
