package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"product/internal/model"
	"product/internal/service"
	"strconv"
	"strings"
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

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	productReq := new(model.CreateProductRequest)

	err := json.NewDecoder(r.Body).Decode(productReq)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.s.CreateProduct(*productReq)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	product, err := h.s.GetProduct(id)
	if err != nil {
		h.l.Error("failed to get product", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		h.l.Error("failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	if err := h.s.DeleteProduct(id); err != nil {
		h.l.Error("failed to delete product", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Get IDs from query parameters
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		http.Error(w, "ids parameter is required", http.StatusBadRequest)
		return
	}

	// Parse the comma-separated IDs
	var ids []int
	for _, idStr := range strings.Split(idsParam, ",") {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id in ids parameter", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	// Get products from service
	products, err := h.s.GetProducts(ids)
	if err != nil {
		h.l.Error("failed to get products", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return products as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		h.l.Error("failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
