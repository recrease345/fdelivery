package api

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CreateOrderReq struct {
	UserID          uuid.UUID       `json:"user_id"`
	RestarauntID    uuid.UUID       `json:"restaurant_id"`
	Items           json.RawMessage `json:"items"`
	DeliveryAddress string          `json:"delivery_address"`
	IdempotencyKey  string          `json:"idempotency_key"`
}
