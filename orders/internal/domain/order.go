package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Created    Status = "created"
	Paid       Status = "paid"
	Preparing  Status = "preparing"
	Ready      Status = "ready"
	Assigning  Status = "assigning"
	Assigned   Status = "assigned"
	Delivering Status = "delivering"
	Delivered  Status = "delivered"
	Canceled   Status = "canceled"
)

type Item struct {
	DishID    uuid.UUID
	Name      string
	UnitPrice int64
	Quantity  int
}

var (
	ErrInvalidUserID       = errors.New("order: userID is nil")
	ErrInvalidItems        = errors.New("order: items mustnt be empty")
	ErrInvalidRestaurantID = errors.New("order: restaurantID is nil")
	ErrInvalidTotalPrice   = errors.New("order: total mismatch")
)

type InvalidTransitionError struct {
	From, To Status
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("order: invalid transition %s -> %s", e.From, e.To)
}

type Order struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RestaurantID uuid.UUID
	Items        []Item
	TotalPrice   int64
	Status       Status
	CreatedAt    time.Time
}

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []Item, totalPrice int64) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if restaurantID == uuid.Nil {
		return nil, ErrInvalidRestaurantID
	}

	if len(items) <= 0 {
		return nil, ErrInvalidItems
	}

	var totalItemsPrice int64
	for _, item := range items {
		// Name != "" {} // сделать валидацию айтема
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
		CreatedAt:    time.Now(),
	}, nil
}
