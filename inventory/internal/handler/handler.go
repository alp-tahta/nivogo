package handler

import (
	"encoding/json"
	"inventory/internal/model"
	"inventory/internal/service"
	"log/slog"
	"net/http"
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

func (h *Handler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	inventoryReq := new(model.CreateInventoryRequest)

	err := json.NewDecoder(r.Body).Decode(inventoryReq)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.s.CreateInventory(*inventoryReq)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetQuantityOfAProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	quantity, err := h.s.GetQuantityOfAProduct(id)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(quantity)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h.l.Info("Successfully retrieved quantity of product", "product_id", id, "quantity", quantity.Quantity)
}

func (h *Handler) GetQuantityOfProducts(w http.ResponseWriter, r *http.Request) {
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
	products, err := h.s.GetQuantityOfProducts(ids)
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

func (h *Handler) ReduceQuantityOfAProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	amount := new(model.ReduceQuantityOfAProduct)
	err = json.NewDecoder(r.Body).Decode(amount)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.s.ReduceQuantityOfAProduct(id, *amount)
	if err != nil {
		h.l.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
