package api

import (
	"encoding/json"
	"fdelivery_orders/internal/order"
	"time"

	"github.com/google/uuid"
)

type CreateOrderReq struct {
	RestaurantID    uuid.UUID       `json:"restaurant_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Items           json.RawMessage `json:"items" validate:"required"`
	DeliveryAddress string          `json:"delivery_address" example:"Улица Пушкина, Дом Колотушкина 2" validate:"required"`
	IdempotencyKey  string          `json:"idempotency_key" example:"jg8d30kd4" validate:"required"`
}

type TransitOrderReq struct {
	From order.Status `json:"from" example:"delivering" validate:"required"`
	To   order.Status `json:"to" example:"delivered" validate:"required"`
}

type CreateOrderResp struct {
	OrderID uuid.UUID    `json:"order_id" validate:"required"`
	Status  order.Status `json:"status" example:"created"`
}

type OrderResp struct {
	ID              uuid.UUID    `json:"id"`
	UserID          uuid.UUID    `json:"user_id"`
	RestaurantID    uuid.UUID    `json:"restaurant_id"`
	Items           []order.Item `json:"items"`
	TotalPrice      int64        `json:"total_price"`
	Status          order.Status `json:"status"`
	DeliveryAddress string       `json:"delivery_address"`
	CreatedAt       time.Time    `json:"created_at"`
}
