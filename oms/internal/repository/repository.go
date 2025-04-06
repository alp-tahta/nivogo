package repository

import (
	"database/sql"
	"log/slog"
	"oms/internal/model"
)

// RepositoryI defines the interface for repository operations
type RepositoryI interface {
	// Order operations
	CreateOrder(order model.CreateOrder) (int, error)
	GetOrder(id int) (*model.Order, error)
	GetOrders() ([]model.Order, error)
	CreateOrderItems(orderID int, items []model.OrderItem) error
	GetOrderItems() (map[int][]model.OrderItem, error)

	// Saga operations
	CreateSaga(saga model.OrderSaga) error
	GetSaga(orderID int) (*model.OrderSaga, error)
	GetSagasByStatus(status string) ([]*model.OrderSaga, error)
	UpdateSagaStatus(orderID int, status string) error
}

// Repository implements the RepositoryI interface
type Repository struct {
	l  *slog.Logger
	db *sql.DB
}

// New creates a new repository instance
func New(l *slog.Logger, db *sql.DB) *Repository {
	return &Repository{
		l:  l,
		db: db,
	}
}

// CreateOrder creates a new order
func (r *Repository) CreateOrder(order model.CreateOrder) (int, error) {
	var orderID int
	err := r.db.QueryRow(
		"INSERT INTO orders (status, created_at) VALUES ($1, $2) RETURNING id",
		order.Status, order.CreatedAt,
	).Scan(&orderID)

	if err != nil {
		r.l.Error("failed to create order", "error", err)
		return 0, err
	}

	return orderID, nil
}

// GetOrder gets an order by ID
func (r *Repository) GetOrder(id int) (*model.Order, error) {
	var order model.Order
	err := r.db.QueryRow("SELECT id, status, created_at FROM orders WHERE id = $1", id).Scan(&order.ID, &order.Status, &order.CreatedAt)
	if err != nil {
		r.l.Error("failed to get order", "error", err, "order_id", id)
		return nil, err
	}
	return &order, nil
}

// GetOrders gets all orders
func (r *Repository) GetOrders() ([]model.Order, error) {
	rows, err := r.db.Query("SELECT id, status, created_at FROM orders")
	if err != nil {
		r.l.Error("failed to get orders", "error", err)
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.ID, &order.Status, &order.CreatedAt)
		if err != nil {
			r.l.Error("failed to scan order", "error", err)
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// CreateOrderItems creates order items for an order
func (r *Repository) CreateOrderItems(orderID int, items []model.OrderItem) error {
	for _, item := range items {
		_, err := r.db.Exec(
			"INSERT INTO order_items (order_id, product_id, product_name, product_description, quantity) VALUES ($1, $2, $3, $4, $5)",
			orderID, item.Product.ID, item.Product.Name, item.Product.Description, item.Quantity,
		)
		if err != nil {
			r.l.Error("failed to create order item", "error", err, "order_id", orderID, "product_id", item.Product.ID)
			return err
		}
	}
	return nil
}

// GetOrderItems gets all order items
func (r *Repository) GetOrderItems() (map[int][]model.OrderItem, error) {
	rows, err := r.db.Query("SELECT order_id, product_id, product_name, product_description, quantity FROM order_items")
	if err != nil {
		r.l.Error("failed to get order items", "error", err)
		return nil, err
	}
	defer rows.Close()

	orderItems := make(map[int][]model.OrderItem)
	for rows.Next() {
		var orderID int
		var item model.OrderItem
		err := rows.Scan(&orderID, &item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Quantity)
		if err != nil {
			r.l.Error("failed to scan order item", "error", err)
			return nil, err
		}
		orderItems[orderID] = append(orderItems[orderID], item)
	}
	return orderItems, nil
}

// CreateSaga creates a new saga
func (r *Repository) CreateSaga(saga model.OrderSaga) error {
	_, err := r.db.Exec(
		"INSERT INTO order_sagas (order_id, status, step, created_at) VALUES ($1, $2, $3, $4)",
		saga.OrderID, saga.Status, saga.Step, saga.CreatedAt,
	)
	if err != nil {
		r.l.Error("failed to create saga", "error", err, "order_id", saga.OrderID)
		return err
	}
	return nil
}

// GetSaga gets a saga by order ID
func (r *Repository) GetSaga(orderID int) (*model.OrderSaga, error) {
	var saga model.OrderSaga
	err := r.db.QueryRow("SELECT order_id, status, step, created_at FROM order_sagas WHERE order_id = $1", orderID).Scan(&saga.OrderID, &saga.Status, &saga.Step, &saga.CreatedAt)
	if err != nil {
		r.l.Error("failed to get saga", "error", err, "order_id", orderID)
		return nil, err
	}
	return &saga, nil
}

// GetSagasByStatus gets all sagas with a specific status
func (r *Repository) GetSagasByStatus(status string) ([]*model.OrderSaga, error) {
	rows, err := r.db.Query("SELECT order_id, status, step, created_at FROM order_sagas WHERE status = $1", status)
	if err != nil {
		r.l.Error("failed to get sagas", "error", err)
		return nil, err
	}
	defer rows.Close()

	var sagas []*model.OrderSaga
	for rows.Next() {
		var saga model.OrderSaga
		err := rows.Scan(&saga.OrderID, &saga.Status, &saga.Step, &saga.CreatedAt)
		if err != nil {
			r.l.Error("failed to scan saga", "error", err)
			return nil, err
		}
		sagas = append(sagas, &saga)
	}
	return sagas, nil
}

// UpdateSagaStatus updates a saga's status
func (r *Repository) UpdateSagaStatus(orderID int, status string) error {
	_, err := r.db.Exec("UPDATE order_sagas SET status = $1 WHERE order_id = $2", status, orderID)
	if err != nil {
		r.l.Error("failed to update saga status", "error", err, "order_id", orderID)
		return err
	}
	return nil
}
