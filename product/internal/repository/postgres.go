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

func (r *Repository) GetProduct(id int) (*model.Product, error) {
	query := `SELECT id, name, description, price FROM products WHERE id = $1`
	product := &model.Product{}
	err := r.db.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("could not get product: %w", err)
	}
	return product, nil
}

func (r *Repository) UpdateProduct(id int, req model.CreateProductRequest) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3 WHERE id = $4`
	result, err := r.db.Exec(query, req.Name, req.Description, req.Price, id)
	if err != nil {
		return fmt.Errorf("could not update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *Repository) DeleteProduct(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("could not delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
