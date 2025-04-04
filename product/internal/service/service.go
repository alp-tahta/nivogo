package service

import (
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
	GetProduct()
	UpdateProduct()
	DeleteProduct()
}

func (s *Service) CreateProduct(req model.CreateProductRequest) error {
	err := s.r.CreateProduct(req)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProduct() {}

func (s *Service) UpdateProduct() {}

func (s *Service) DeleteProduct() {}
