package api

import (
	"errors"
	"fdelivery_orders/internal/order"
	"log/slog"
	"net/http"
)

type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteErr(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error, req any) {
	var tr *order.InvalidTransitionError

	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		writeJSONError(w, http.StatusNotFound, errBody{Code: "order_not_found", Message: err.Error()})

	case errors.As(err, &tr):
		writeJSONError(w, http.StatusConflict, errBody{Code: "invalid_transition", Message: err.Error()})

	case errors.Is(err, order.ErrValidation):
		writeJSONError(w, http.StatusBadRequest, errBody{Code: "validation_failed", Message: err.Error()})

	case errors.Is(err, order.ErrDuplicateIdempotencyKey):
		writeJSONError(w, http.StatusConflict, errBody{Code: "idempotency_conflict", Message: err.Error()})

	case errors.Is(err, order.ErrEmptyKey):
		writeJSONError(w, http.StatusConflict, errBody{Code: "idempotency_key_required", Message: err.Error()})

	default:
		log.Error("unmapped error", "req", req, "err", err)
		writeJSONError(w, http.StatusInternalServerError, errBody{Code: "internal", Message: "internal error"})
	}
}
