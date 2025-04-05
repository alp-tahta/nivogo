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
