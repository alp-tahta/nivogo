package repository

import "product/internal/model"

type RepositoryI interface {
	CreateProduct(req model.CreateProductRequest) error
	GetProduct(id int) (*model.Product, error)
	GetProducts(ids []int) ([]model.Product, error)
	DeleteProduct(id int) error
}
