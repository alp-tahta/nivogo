package routes

import (
	"net/http"
	"oms/internal/handler"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /orders", h.CreateOrder)
	mux.HandleFunc("GET /orders", h.GetOrders)
}
