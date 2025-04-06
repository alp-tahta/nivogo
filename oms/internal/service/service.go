package service

import (
	"fmt"
	"log/slog"
	"oms/internal/kafka"
	"oms/internal/model"
	"oms/internal/repository"
	"time"
)

type Service struct {
	l *slog.Logger
	r repository.RepositoryI
	k *kafka.KafkaClient
}

func New(l *slog.Logger, r repository.RepositoryI) (*Service, error) {
	k, err := kafka.New(l)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Service{
		l: l,
		r: r,
		k: k,
	}, nil
}

type ServiceI interface {
	CreateOrder(order model.CreateOrder) (int, error)
	CreateOrderFromRequest(req model.CreateOrderRequest) error
	GetOrders() ([]model.Order, error)
	Close() error
}

func (s *Service) Close() error {
	return s.k.Close()
}

func (s *Service) CreateOrderFromRequest(req model.CreateOrderRequest) error {
	// Validate request
	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	// Create order from request
	order := model.CreateOrder{
		Items:     req.Items,
		Status:    "CREATED",
		CreatedAt: time.Now(),
	}

	// Create order
	_, err := s.CreateOrder(order)
	return err
}

func (s *Service) CreateOrder(order model.CreateOrder) (int, error) {
	// Create order in database and get the ID
	orderID, err := s.r.CreateOrder(order)
	if err != nil {
		s.l.Error("failed to create order", "error", err)
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	// Start saga
	saga := model.OrderSaga{
		OrderID:   orderID,
		Status:    "STARTED",
		Step:      0,
		CreatedAt: time.Now(),
	}

	// Save saga state
	err = s.r.CreateSaga(saga)
	if err != nil {
		s.l.Error("failed to create saga", "error", err, "order_id", orderID)
		return orderID, fmt.Errorf("failed to create saga: %w", err)
	}

	// Step 1: Reserve inventory for all items
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return orderID, fmt.Errorf("invalid quantity for product %d", item.Product.ID)
		}

		if err := s.reserveInventory(item.Product.ID, item.Quantity); err != nil {
			// Compensating transaction: Release all previously reserved inventory
			for _, releasedItem := range order.Items {
				if releasedItem.Product.ID == item.Product.ID {
					break // Skip the current item as it wasn't reserved
				}
				s.releaseInventory(releasedItem.Product.ID, releasedItem.Quantity)
			}
			s.r.UpdateSagaStatus(saga.OrderID, "FAILED")
			return orderID, fmt.Errorf("failed to reserve inventory for product %d: %w", item.Product.ID, err)
		}
	}

	// Update saga step
	s.r.UpdateSagaStatus(saga.OrderID, "INVENTORY_RESERVED")

	// Step 2: Create order items
	err = s.r.CreateOrderItems(orderID, order.Items)
	if err != nil {
		// Compensating transaction: Release all inventory
		for _, item := range order.Items {
			s.releaseInventory(item.Product.ID, item.Quantity)
		}
		s.r.UpdateSagaStatus(saga.OrderID, "FAILED")
		return orderID, fmt.Errorf("failed to create order items: %w", err)
	}

	// Update saga status to completed
	s.r.UpdateSagaStatus(saga.OrderID, "COMPLETED")

	return orderID, nil
}

func (s *Service) GetOrders() ([]model.Order, error) {
	orders, err := s.r.GetOrders()
	if err != nil {
		s.l.Error("failed to get orders", "error", err)
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orderItems, err := s.r.GetOrderItems()
	if err != nil {
		s.l.Error("failed to get order items", "error", err)
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	// Combine orders with their items
	for i := range orders {
		if items, exists := orderItems[orders[i].ID]; exists {
			orders[i].Items = items
		}
	}

	return orders, nil
}

func (s *Service) reserveInventory(productID int, quantity int) error {
	// Send reserve inventory request via Kafka
	err := s.k.ReserveInventory(productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to send reserve inventory request: %w", err)
	}

	// Wait for response with timeout
	err = s.k.WaitForInventoryResponse(productID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to reserve inventory: %w", err)
	}

	return nil
}

func (s *Service) releaseInventory(productID int, quantity int) error {
	// Send release inventory request via Kafka
	err := s.k.ReleaseInventory(productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to send release inventory request: %w", err)
	}

	// Wait for response with timeout
	err = s.k.WaitForInventoryResponse(productID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to release inventory: %w", err)
	}

	return nil
}
