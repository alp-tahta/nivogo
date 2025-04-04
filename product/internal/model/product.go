package model

type Product struct {
	ID          int
	Name        string
	Description string
	Price       int
}

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type UpdateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}
