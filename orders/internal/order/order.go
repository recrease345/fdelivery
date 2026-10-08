package order

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID          = errors.New("order: userID is nil")
	ErrInvalidItems           = errors.New("order: items must be non-empty")
	ErrInvalidRestaurantID    = errors.New("order: restaurantID is nil")
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

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []Item, deliveryAddress string, now time.Time) (*Order, error) {
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

	var totalPrice int64
	for _, item := range items {
		if item.DishID == uuid.Nil || item.Name == "" || item.Quantity <= 0 || item.UnitPrice <= 0 {
			return nil, ErrInvalidItemsAttributes
		}

		totalPrice += (item.UnitPrice * int64(item.Quantity))
	}

	return &Order{
		ID:              uuid.New(),
		UserID:          userID,
		RestaurantID:    restaurantID,
		Items:           slices.Clone(items),
		TotalPrice:      totalPrice,
		Status:          Created,
		DeliveryAddress: deliveryAddress,
		CreatedAt:       now,
	}, nil
}
