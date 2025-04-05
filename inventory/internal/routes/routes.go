package routes

import (
	"inventory/internal/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /inventory", h.CreateInventory)
	mux.HandleFunc("GET /inventory/{id}", h.GetQuantityOfAProduct)
}
