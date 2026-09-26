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

var (
	ErrInvalidUserID = errors.New("order: userID is nil")
	ErrInvalidItems = errors.New("order: items mustnt be empty")
	ErrInvalidRestaurantID = errors.New("order: restaurantID is nil")
)

type Order struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RestaurantID uuid.UUID
	Items        []uuid.UUID
	Status       Status
	CreatedAt    *time.Time
}

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []uuid.UUID, status Status) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if restaurantID == uuid.Nil {
		return nil, ErrInvalidRestaurantID
	}

	if len(items) <= 0 {
		return nil, ErrInvalidItems
	}

	return &Order{
		ID: uuid.New(),
		UserID: userID,
		RestaurantID: restaurantID,
		Items: items,
		Status: status,
	}, nil
}
