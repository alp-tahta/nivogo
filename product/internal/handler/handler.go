package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"product/internal/model"
	"product/internal/service"
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

}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {

}
