package model

type Inventory struct {
	ProductID int
	Quantity  int
}

type CreateInventoryRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type QuantityOfAProduct struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ReduceQuantityOfAProduct struct {
	Amount int `json:"amount"`
}
