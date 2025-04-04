package repository

import "product/internal/model"

type RepositoryI interface {
	CreateProduct(req model.CreateProductRequest) error
	GetProduct()
	UpdateProduct()
	DeleteProduct()
}
