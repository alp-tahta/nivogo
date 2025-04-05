package routes

import (
	"net/http"
	"product/internal/handler"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func RegisterRoutes(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("GET /health", corsMiddleware(h.Health))
	mux.HandleFunc("POST /product", corsMiddleware(h.CreateProduct))
	mux.HandleFunc("GET /product/{id}", corsMiddleware(h.GetProduct))
	mux.HandleFunc("GET /product", corsMiddleware(h.GetProducts))
	mux.HandleFunc("DELETE /product/{id}", corsMiddleware(h.DeleteProduct))
}
