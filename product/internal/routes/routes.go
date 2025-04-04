package routes

import (
	"net/http"
	"product/internal/handler"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /product", h.CreateProduct)
	mux.HandleFunc("GET /product/{id}", h.GetProduct)
	mux.HandleFunc("GET /product", h.GetProducts)
	mux.HandleFunc("DELETE /product/{id}", h.DeleteProduct)
}
