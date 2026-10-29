package api

import (
	"encoding/json"
	"fdelivery_orders/internal/order"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	Service order.Service
	log     *slog.Logger
}

const MaxBytesReader = 1048576 // 1mb

func checkContentType(w http.ResponseWriter, r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytesReader)

	if !checkContentType(w, r) {
		writeJSONError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}

	var req CreateOrderReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	var items []order.Item

	if err := json.Unmarshal(req.Items, &items); err != nil {
		writeJSONError(w, http.StatusBadRequest, "items required")
		return
	}

	cmd := order.CreateCommand{
		UserID:          req.UserID,
		RestarauntID:    req.RestarauntID,
		Items:           items,
		DeliveryAddress: req.DeliveryAddress,
		IdempotencyKey:  req.IdempotencyKey,
	}

	res, err := h.Service.Create(r.Context(), cmd)
	if err != nil {
		WriteErr(w, r, h.log, err, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write(res.ResponseBody)
}
