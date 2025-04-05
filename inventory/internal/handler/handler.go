package handler

import (
	"encoding/json"
	"fmt"
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

func (h *Handler) ReserveInventory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.l.Error("invalid product id", "error", err)
		sendError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Geçersiz ürün ID'si", "Ürün ID'si sayısal bir değer olmalıdır")
		return
	}

	var req model.ReserveInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error("invalid request body", "error", err)
		sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", "İstek gövdesi JSON formatında olmalıdır")
		return
	}

	if req.Quantity <= 0 {
		h.l.Error("invalid quantity", "quantity", req.Quantity)
		sendError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Geçersiz miktar", "Miktar 0'dan büyük olmalıdır")
		return
	}

	err = h.s.ReserveInventory(id, req)
	if err != nil {
		h.l.Error("failed to reserve inventory", "error", err, "product_id", id)
		if err.Error() == "not enough quantity available" {
			sendError(w, http.StatusBadRequest, "INSUFFICIENT_QUANTITY", "Yetersiz stok",
				fmt.Sprintf("Ürün ID: %d için yeterli stok bulunmamaktadır", id))
			return
		}
		sendError(w, http.StatusInternalServerError, "RESERVE_FAILED", "Stok rezervasyonu başarısız",
			fmt.Sprintf("Ürün ID: %d için stok rezerve edilemedi", id))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.l.Info("inventory reserved successfully", "product_id", id, "quantity", req.Quantity)
}

func (h *Handler) ReleaseInventory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.l.Error("invalid product id", "error", err)
		sendError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Geçersiz ürün ID'si", "Ürün ID'si sayısal bir değer olmalıdır")
		return
	}

	var req model.ReleaseInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error("invalid request body", "error", err)
		sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", "İstek gövdesi JSON formatında olmalıdır")
		return
	}

	if req.Quantity <= 0 {
		h.l.Error("invalid quantity", "quantity", req.Quantity)
		sendError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Geçersiz miktar", "Miktar 0'dan büyük olmalıdır")
		return
	}

	err = h.s.ReleaseInventory(id, req)
	if err != nil {
		h.l.Error("failed to release inventory", "error", err, "product_id", id)
		sendError(w, http.StatusInternalServerError, "RELEASE_FAILED", "Stok serbest bırakma başarısız",
			fmt.Sprintf("Ürün ID: %d için stok serbest bırakılamadı", id))
		return
	}

	w.WriteHeader(http.StatusOK)
	h.l.Info("inventory released successfully", "product_id", id, "quantity", req.Quantity)
}

func sendError(w http.ResponseWriter, status int, code string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}
