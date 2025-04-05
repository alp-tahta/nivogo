package repository

import (
	"database/sql"
	"fmt"
	"inventory/internal/model"
	"log/slog"
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

func (r *Repository) CreateInventory(req model.CreateInventoryRequest) error {
	query := `INSERT INTO inventory (product_id, quantity) VALUES ($1, $2)`
	_, err := r.db.Exec(query, req.ProductID, req.Quantity)
	if err != nil {
		return fmt.Errorf("could not insert inventory: %w", err)
	}

	return nil
}

func (r *Repository) GetQuantityOfAProduct(id int) (*model.QuantityOfAProduct, error) {
	query := `SELECT product_id, quantity FROM inventory WHERE id = $1`
	quantity := &model.QuantityOfAProduct{}
	err := r.db.QueryRow(query, id).Scan(&quantity.ProductID, &quantity.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("could not get product: %w", err)
	}
	return quantity, nil
}
