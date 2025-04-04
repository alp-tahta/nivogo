package service

import (
	"fmt"
	"log/slog"
	"product/internal/model"
	"product/internal/repository"
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
	CreateProduct(req model.CreateProductRequest) error
	GetProduct(id int) (*model.Product, error)
	GetProducts(ids []int) ([]model.Product, error)
	DeleteProduct(id int) error
}

func (s *Service) CreateProduct(req model.CreateProductRequest) error {
	s.l.Info("creating new product", "name", req.Name)

	err := s.r.CreateProduct(req)
	if err != nil {
		s.l.Error("failed to create product", "error", err)
		return fmt.Errorf("failed to create product: %w", err)
	}

	s.l.Info("product created successfully", "name", req.Name)
	return nil
}

func (s *Service) GetProduct(id int) (*model.Product, error) {
	s.l.Info("getting product", "id", id)

	product, err := s.r.GetProduct(id)
	if err != nil {
		s.l.Error("failed to get product", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	s.l.Info("product retrieved successfully", "id", id)
	return product, nil
}

func (s *Service) DeleteProduct(id int) error {
	s.l.Info("deleting product", "id", id)

	err := s.r.DeleteProduct(id)
	if err != nil {
		s.l.Error("failed to delete product", "id", id, "error", err)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	s.l.Info("product deleted successfully", "id", id)
	return nil
}

func (s *Service) GetProducts(ids []int) ([]model.Product, error) {
	s.l.Info("getting products", "ids", ids)

	products, err := s.r.GetProducts(ids)
	if err != nil {
		s.l.Error("failed to get products", "ids", ids, "error", err)
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	s.l.Info("products retrieved successfully", "count", len(products))
	return products, nil
}
