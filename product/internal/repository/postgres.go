package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"product/internal/model"
)

type Repository struct {
	l  *slog.Logger
	db *sql.DB
}

func New(l *slog.Logger, db *sql.DB) *Repository {
	return &Repository{
		l:  l,
		db: db,
	}
}

func (r *Repository) CreateProduct(req model.CreateProductRequest) error {
	query := `INSERT INTO products (name, description, price) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, req.Name, req.Description, req.Price)
	if err != nil {
		return fmt.Errorf("could not insert product: %w", err)
	}

	return nil
}

func (r *Repository) GetProduct() {}

func (r *Repository) UpdateProduct() {}

func (r *Repository) DeleteProduct() {}
