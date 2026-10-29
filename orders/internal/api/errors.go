package api

import (
	"errors"
	"fdelivery_orders/internal/order"
	"log/slog"
	"net/http"
)

type ErrBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteErr(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error, req any) {
	var tr *order.InvalidTransitionError

	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		writeJSONError(w, http.StatusNotFound, ErrBody{Code: "order_not_found", Message: err.Error()})

	case errors.As(err, &tr):
		writeJSONError(w, http.StatusConflict, ErrBody{Code: "invalid_transition", Message: err.Error()})

	case errors.Is(err, order.ErrValidation):
		writeJSONError(w, http.StatusBadRequest, ErrBody{Code: "validation_failed", Message: err.Error()})

	case errors.Is(err, order.ErrDuplicateIdempotencyKey):
		writeJSONError(w, http.StatusConflict, ErrBody{Code: "idempotency_conflict", Message: err.Error()})

	case errors.Is(err, order.ErrEmptyKey):
		writeJSONError(w, http.StatusBadRequest, ErrBody{Code: "idempotency_key_required", Message: err.Error()})

	default:
		log.Error("unmapped error", "req", req, "err", err)
		writeJSONError(w, http.StatusInternalServerError, ErrBody{Code: "internal", Message: "internal error"})
	}
}
