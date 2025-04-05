package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"oms/internal/model"
	"oms/internal/service"
)

type Handler struct {
	l *slog.Logger
	s service.ServiceI
}

func New(l *slog.Logger, s service.ServiceI) *Handler {
	return &Handler{
		l: l,
		s: s,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate order
	if len(order.Items) == 0 {
		h.sendError(w, http.StatusBadRequest, "Order must contain at least one item", nil)
		return
	}

	// Create order
	if err := h.s.CreateOrder(order); err != nil {
		h.l.Error("failed to create order", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.s.GetOrders()
	if err != nil {
		h.l.Error("failed to get orders", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to get orders", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string, details interface{}) {
	var detailsStr string
	if details != nil {
		if str, ok := details.(string); ok {
			detailsStr = str
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   message,
		Code:    fmt.Sprintf("ERR_%d", status),
		Details: detailsStr,
	})
}
