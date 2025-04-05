package repository

import "inventory/internal/model"

type RepositoryI interface {
	CreateInventory(req model.CreateInventoryRequest) error
	GetQuantityOfAProduct(id int) (*model.QuantityOfAProduct, error)
	GetQuantityOfProducts(ids []int) ([]model.Inventory, error)
}
