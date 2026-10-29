package api

import (
	"encoding/json"
	"fdelivery_orders/internal/order"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	Service order.Service
	log     *slog.Logger
}

func NewHandler(svc *order.Service, log *slog.Logger) *Handler {
	return &Handler{Service: *svc, log: log}
}

const MaxBytesReader = 1048576 // 1mb

func checkContentType(w http.ResponseWriter, r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

// @Summary CreateOrder
// @Description Create new order
// @Tags Orders
// @Accept json
// @Produce json
// @Param input body CreateOrderReq true "Order create data"
// @Param X-User-ID header string true "ID юзера"
// @Success 202 {object} CreateOrderResp
// @Failure 415 {object} ErrBody
// @Failure 400 {object} ErrBody
// @Failure 401 {object} ErrBody
// @Router /orders [post]
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

	headerUserID := r.Header.Get("X-User-ID")
	if headerUserID == "" {
		writeJSONError(w, http.StatusUnauthorized, "user id is empty")
		return
	}

	userID, err := uuid.Parse(headerUserID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	cmd := order.CreateCommand{
		UserID:          userID,
		RestaurantID:    req.RestaurantID,
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	QueryID := r.PathValue("id")
	if QueryID == "" {
		writeJSONError(w, http.StatusBadRequest, "query parameter required")
		return
	}

	id, err := uuid.Parse(QueryID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	headerUserID := r.Header.Get("X-User-ID")
	if headerUserID == "" {
		writeJSONError(w, http.StatusUnauthorized, "user id is empty")
		return
	}

	userID, err := uuid.Parse(headerUserID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	o, err := h.Service.Get(r.Context(), id, userID)
	if err != nil {
		WriteErr(w, r, h.log, err, map[string]any{
			"order_id": id,
			"user_id":  userID,
		})

		return
	}

	writeJSON(w, http.StatusOK, &OrderResp{
		ID:              o.ID(),
		UserID:          o.UserID(),
		RestaurantID:    o.RestaurantID(),
		Items:           o.Items(),
		TotalPrice:      o.TotalPrice(),
		Status:          o.Status(),
		DeliveryAddress: o.DeliveryAddress(),
		CreatedAt:       o.CreatedAt(),
	})
}

func (h *Handler) Transit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytesReader)

	if !checkContentType(w, r) {
		writeJSONError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}

	queryID := r.PathValue("id")
	if queryID == "" {
		writeJSONError(w, http.StatusBadRequest, "query parameter required")
		return
	}

	id, err := uuid.Parse(queryID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var req TransitOrderReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.Service.Transit(r.Context(), id, req.From, req.To); err != nil {
		WriteErr(w, r, h.log, err, req)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytesReader)

	if !checkContentType(w, r) {
		writeJSONError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}

	QueryID := r.PathValue("id")
	if QueryID == "" {
		writeJSONError(w, http.StatusBadRequest, "query parameter required")
		return
	}

	id, err := uuid.Parse(QueryID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	headerUserID := r.Header.Get("X-User-ID")
	if headerUserID == "" {
		writeJSONError(w, http.StatusUnauthorized, "user id is empty")
		return
	}

	userID, err := uuid.Parse(headerUserID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	_, err = h.Service.Cancel(r.Context(), id, userID)
	if err != nil {
		WriteErr(w, r, h.log, err, nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]order.Status{
		"status": order.Cancelled,
	})
}
