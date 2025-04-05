package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"oms/internal/model"
	"oms/internal/service"
	"time"
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

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		h.l.Error("Failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Create order using service
	if err := h.s.CreateOrderFromRequest(req); err != nil {
		h.l.Error("failed to create order", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to create order", err.Error())
		return
	}

	// Return only status code, no response body
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

// Helper function to generate a unique order ID
func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}
