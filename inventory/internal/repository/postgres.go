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
	query := `SELECT product_id, quantity FROM inventory WHERE product_id = $1`
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

func (r *Repository) GetQuantityOfProducts(ids []int) ([]model.Inventory, error) {
	if len(ids) == 0 {
		return []model.Inventory{}, nil
	}

	// Create a parameterized query with the correct number of placeholders
	query := `SELECT product_id, quantity FROM inventory WHERE product_id IN (`
	for i := range ids {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+1)
	}
	query += ")"

	// Convert ids slice to interface{} slice for Exec
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query inventory: %w", err)
	}
	defer rows.Close()

	// Process the results
	var products []model.Inventory
	for rows.Next() {
		var product model.Inventory
		err := rows.Scan(&product.ProductID, &product.Quantity)
		if err != nil {
			return nil, fmt.Errorf("could not scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating Inventory rows: %w", err)
	}

	return products, nil
}
