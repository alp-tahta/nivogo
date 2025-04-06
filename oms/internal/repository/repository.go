package repository

import (
	"database/sql"
	"log/slog"
	"oms/internal/model"
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

type RepositoryI interface {
	CreateOrder(order model.CreateOrder) (int, error)
	CreateOrderItems(orderID int, items []model.OrderItem) error
	GetOrders() ([]model.Order, error)
	GetOrderItems() (map[int][]model.OrderItem, error)
	CreateSaga(saga model.OrderSaga) error
	UpdateSagaStatus(orderID int, status string) error
}

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

func (r *Repository) UpdateSagaStatus(orderID int, status string) error {
	_, err := r.db.Exec("UPDATE order_sagas SET status = $1 WHERE order_id = $2", status, orderID)
	if err != nil {
		r.l.Error("failed to update saga status", "error", err, "order_id", orderID)
		return err
	}
	return nil
}
