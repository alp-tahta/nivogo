package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"oms/internal/model"
	"oms/internal/repository"
	"time"
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
	CreateOrder(order model.CreateOrder) error
	CreateOrderFromRequest(req model.CreateOrderRequest) error
	GetOrders() ([]model.Order, error)
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
	return s.CreateOrder(order)
}

func (s *Service) CreateOrder(order model.CreateOrder) error {
	// Create order in database
	err := s.r.CreateOrder(order)
	if err != nil {
		s.l.Error("failed to create order", "error", err)
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Get the created order to get its ID
	orders, err := s.r.GetOrders()
	if err != nil {
		s.l.Error("failed to get orders", "error", err)
		return fmt.Errorf("failed to get orders: %w", err)
	}

	// Find the most recently created order
	if len(orders) == 0 {
		return fmt.Errorf("failed to find created order")
	}
	createdOrder := orders[len(orders)-1]

	// Start saga
	saga := model.OrderSaga{
		OrderID:   createdOrder.ID,
		Status:    "STARTED",
		Step:      0,
		CreatedAt: time.Now(),
	}

	// Save saga state
	err = s.r.CreateSaga(saga)
	if err != nil {
		s.l.Error("failed to create saga", "error", err, "order_id", createdOrder.ID)
		return fmt.Errorf("failed to create saga: %w", err)
	}

	// Step 1: Reserve inventory for all items
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity for product %d", item.Product.ID)
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
			return fmt.Errorf("failed to reserve inventory for product %d: %w", item.Product.ID, err)
		}
	}

	// Update saga step
	s.r.UpdateSagaStatus(saga.OrderID, "INVENTORY_RESERVED")

	// Step 2: Create order items
	err = s.r.CreateOrderItems(createdOrder.ID, order.Items)
	if err != nil {
		// Compensating transaction: Release all inventory
		for _, item := range order.Items {
			s.releaseInventory(item.Product.ID, item.Quantity)
		}
		s.r.UpdateSagaStatus(saga.OrderID, "FAILED")
		return fmt.Errorf("failed to create order items: %w", err)
	}

	// Update saga status to completed
	s.r.UpdateSagaStatus(saga.OrderID, "COMPLETED")

	return nil
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
	// Create request body with quantity
	requestBody := model.ReserveInventoryRequest{
		Quantity: quantity,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Call inventory service to reserve items
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:9081/reserve/%d", productID),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("inventory service error: %s - %s", errorResp.Error, errorResp.Details)
		}
		return fmt.Errorf("inventory service error: %s", resp.Status)
	}
	return nil
}

func (s *Service) releaseInventory(productID int, quantity int) error {
	// Create request body with quantity
	requestBody := model.ReleaseInventoryRequest{
		Quantity: quantity,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Call inventory service to release items
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:9081/release/%d", productID),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp model.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("inventory service error: %s - %s", errorResp.Error, errorResp.Details)
		}
		return fmt.Errorf("inventory service error: %s", resp.Status)
	}
	return nil
}
