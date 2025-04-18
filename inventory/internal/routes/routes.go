package routes

import (
	"inventory/internal/handler"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /inventory", h.CreateInventory)
	mux.HandleFunc("GET /inventory/{id}", h.GetQuantityOfAProduct)
	mux.HandleFunc("GET /inventory", h.GetQuantityOfProducts)
	mux.HandleFunc("POST /inventory/product/{id}/reduce", h.ReduceQuantityOfAProduct)
}
