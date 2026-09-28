package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID          = errors.New("order: userID is nil")
	ErrInvalidItems           = errors.New("order: items mustnt be empty")
	ErrInvalidRestaurantID    = errors.New("order: restaurantID is nil")
	ErrInvalidTotalPrice      = errors.New("order: total mismatch")
	ErrInvalidItemsAttributes = errors.New("order: invalid item attributes")
	ErrInvalidDeliveryAddress = errors.New("order: invalid delivery address")
)

type Item struct {
	DishID    uuid.UUID
	Name      string
	UnitPrice int64
	Quantity  int
}

type Order struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	RestaurantID    uuid.UUID
	Items           []Item
	TotalPrice      int64
	Status          Status
	DeliveryAddress string
	CreatedAt       time.Time
}

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []Item, totalPrice int64, deliveryAddress string, now time.Time) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if restaurantID == uuid.Nil {
		return nil, ErrInvalidRestaurantID
	}

	if len(items) <= 0 {
		return nil, ErrInvalidItems
	}

	if deliveryAddress == "" {
		return nil, ErrInvalidDeliveryAddress
	}

	var totalItemsPrice int64
	for _, item := range items {
		if item.DishID == uuid.Nil || item.Name == "" || item.Quantity <= 0 || item.UnitPrice <= 0 {
			return nil, ErrInvalidItemsAttributes
		}

		totalItemsPrice += (item.UnitPrice * int64(item.Quantity))
	}

	if totalItemsPrice != totalPrice {
		return nil, ErrInvalidTotalPrice
	}

	return &Order{
		ID:           uuid.New(),
		UserID:       userID,
		RestaurantID: restaurantID,
		Items:        items,
		Status:       Created,
		CreatedAt:    now,
	}, nil
}
