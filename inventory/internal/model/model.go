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
	Quantity int `json:"quantity"`
}

type ReleaseInventoryRequest struct {
	Quantity int `json:"quantity"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}
