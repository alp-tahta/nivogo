package repository

import "product/internal/model"

type RepositoryI interface {
	CreateProduct(req model.CreateProductRequest) error
	GetProduct(id int) (*model.Product, error)
	UpdateProduct(id int, req model.CreateProductRequest) error
	DeleteProduct(id int) error
}
