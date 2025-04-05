package routes

import (
	"net/http"
	"oms/internal/handler"
)

func RegisterRoutes(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	// Order endpoints
	mux.HandleFunc("POST /orders", h.CreateOrder)
	mux.HandleFunc("GET /orders", h.GetOrders)

	return mux
}
