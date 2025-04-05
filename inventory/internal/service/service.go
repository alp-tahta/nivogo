package service

import (
	"fmt"
	"inventory/internal/model"
	"inventory/internal/repository"
	"log/slog"
)

type Service struct {
	l *slog.Logger
	r repository.RepositoryI
}

func New(l *slog.Logger, r repository.RepositoryI) *Service {
	return &Service{
		l: l,
		r: r,
	}
}

type ServiceI interface {
	CreateInventory(req model.CreateInventoryRequest) error
	GetQuantityOfAProduct(id int) (*model.QuantityOfAProduct, error)
	GetQuantityOfProducts(ids []int) ([]model.Inventory, error)
	ReduceQuantityOfAProduct(id int, a model.ReduceQuantityOfAProduct) error
	ReserveInventory(id int, req model.ReserveInventoryRequest) error
	ReleaseInventory(id int, req model.ReleaseInventoryRequest) error
}

func (s *Service) CreateInventory(req model.CreateInventoryRequest) error {
	err := s.r.CreateInventory(req)
	if err != nil {
		s.l.Error("failed to create inventoruy for product", "error", err)
		return fmt.Errorf("failed to create inventoruy for product: %w", err)
	}

	return nil
}

func (s *Service) GetQuantityOfAProduct(id int) (*model.QuantityOfAProduct, error) {
	quantity, err := s.r.GetQuantityOfAProduct(id)
	if err != nil {
		s.l.Error("failed to get quantity of product", "error", err, "product_id", id)
		return nil, fmt.Errorf("failed to get quantity of product: %w", err)
	}

	return quantity, nil
}

func (s *Service) GetQuantityOfProducts(ids []int) ([]model.Inventory, error) {
	s.l.Info("getting quantities", "ids", ids)

	products, err := s.r.GetQuantityOfProducts(ids)
	if err != nil {
		s.l.Error("failed to get quantities", "ids", ids, "error", err)
		return nil, fmt.Errorf("failed to get quantities: %w", err)
	}

	s.l.Info("quantities retrieved successfully", "count", len(products))
	return products, nil
}

func (s *Service) ReduceQuantityOfAProduct(id int, a model.ReduceQuantityOfAProduct) error {
	quantity, err := s.r.GetQuantityOfAProduct(id)
	if err != nil {
		s.l.Error("failed to get quantity of product", "error", err, "product_id", id)
		return fmt.Errorf("failed to get quantity of product: %w", err)
	}

	if quantity.Quantity < a.Amount {
		return fmt.Errorf("not enough quantity")
	}

	quantity.Quantity = quantity.Quantity - a.Amount

	err = s.r.UpdateQuantityOfAProduct(id, quantity.Quantity)
	if err != nil {
		s.l.Error("failed to update quantity of product", "error", err, "product_id", id)
		return fmt.Errorf("failed to update quantity of product: %w", err)
	}

	return nil
}

func (s *Service) ReserveInventory(id int, req model.ReserveInventoryRequest) error {
	quantity, err := s.r.GetQuantityOfAProduct(id)
	if err != nil {
		s.l.Error("failed to get quantity of product", "error", err, "product_id", id)
		return fmt.Errorf("failed to get quantity of product: %w", err)
	}

	if quantity.Quantity < req.Quantity {
		return fmt.Errorf("not enough quantity available")
	}

	// Reserve the quantity by reducing it
	quantity.Quantity = quantity.Quantity - req.Quantity

	err = s.r.UpdateQuantityOfAProduct(id, quantity.Quantity)
	if err != nil {
		s.l.Error("failed to reserve inventory", "error", err, "product_id", id)
		return fmt.Errorf("failed to reserve inventory: %w", err)
	}

	s.l.Info("inventory reserved successfully", "product_id", id, "quantity", req.Quantity)
	return nil
}

func (s *Service) ReleaseInventory(id int, req model.ReleaseInventoryRequest) error {
	quantity, err := s.r.GetQuantityOfAProduct(id)
	if err != nil {
		s.l.Error("failed to get quantity of product", "error", err, "product_id", id)
		return fmt.Errorf("failed to get quantity of product: %w", err)
	}

	// Release the quantity by adding it back
	quantity.Quantity = quantity.Quantity + req.Quantity

	err = s.r.UpdateQuantityOfAProduct(id, quantity.Quantity)
	if err != nil {
		s.l.Error("failed to release inventory", "error", err, "product_id", id)
		return fmt.Errorf("failed to release inventory: %w", err)
	}

	s.l.Info("inventory released successfully", "product_id", id, "quantity", req.Quantity)
	return nil
}
