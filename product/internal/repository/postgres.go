package repository

import (
	"database/sql"
	"errors"
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
	var id int64
	err := r.db.QueryRow(query, req.Name, req.Description, req.Price).Scan(&id)
	if err != nil {
		return errors.New("failed to insert product: " + err.Error())
	}

	return nil
}

func (r *Repository) GetProduct() {}

func (r *Repository) UpdateProduct() {}

func (r *Repository) DeleteProduct() {}
