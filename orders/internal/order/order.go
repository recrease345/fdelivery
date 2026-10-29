package order

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("order: validation failed")

	ErrInvalidUserID          = fmt.Errorf("%w: userID is nil", ErrValidation)
	ErrInvalidItems           = fmt.Errorf("%w: items must be non-empty", ErrValidation)
	ErrInvalidRestaurantID    = fmt.Errorf("%w: restaurantID is nil", ErrValidation)
	ErrInvalidItemsAttributes = fmt.Errorf("%w: invalid item attributes", ErrValidation)
	ErrInvalidDeliveryAddress = fmt.Errorf("%w: invalid delivery address", ErrValidation)
	ErrInvalidStatus          = fmt.Errorf("%w: restore with invalid status", ErrValidation)
	ErrInvalidID              = fmt.Errorf("%w: id is nil", ErrValidation)

	ErrOrderNotFound = errors.New("order not found")
)

const maxDeliveryAdressLen = 300

type Item struct {
	DishID    uuid.UUID `json:"dish_id"`
	Name      string    `json:"name"`
	UnitPrice int64     `json:"unit_price"`
	Quantity  int       `json:"quantity"`
}

type Order struct {
	id              uuid.UUID
	userID          uuid.UUID
	restaurantID    uuid.UUID
	items           []Item
	totalPrice      int64
	status          Status
	deliveryAddress string
	createdAt       time.Time
}

func NewOrder(userID uuid.UUID, restaurantID uuid.UUID, items []Item, deliveryAddress string, now time.Time) (*Order, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if restaurantID == uuid.Nil {
		return nil, ErrInvalidRestaurantID
	}

	if len(items) == 0 {
		return nil, ErrInvalidItems
	}

	if deliveryAddress == "" || len(deliveryAddress) > maxDeliveryAdressLen {
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
		id:              uuid.New(),
		userID:          userID,
		restaurantID:    restaurantID,
		items:           slices.Clone(items),
		totalPrice:      totalPrice,
		status:          Created,
		deliveryAddress: deliveryAddress,
		createdAt:       now,
	}, nil
}

func Restore(id uuid.UUID, userID uuid.UUID, restaurantID uuid.UUID, totalPrice int64, items []Item, deliveryAddress string, createdAt time.Time, status Status) (*Order, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}

	switch status {
	case Created, Paid, Preparing, Ready, Assigning, Assigned, Delivering, Delivered, Cancelled, FailedAssign:
	default:
		return nil, ErrInvalidStatus
	}

	return &Order{
		id:              id,
		userID:          userID,
		restaurantID:    restaurantID,
		items:           slices.Clone(items),
		totalPrice:      totalPrice,
		status:          status,
		deliveryAddress: deliveryAddress,
		createdAt:       createdAt,
	}, nil
}

func (o *Order) cancel() (Compensation, error) {
	comp := compensations[transition{o.status, Cancelled}]
	if err := o.transitionTo(Cancelled); err != nil {
		return CompUnset, err
	}

	return comp, nil
}

func (o *Order) paid() error {
	if err := o.transitionTo(Paid); err != nil {
		return err
	}

	return nil
}

func (o *Order) preparing() error {
	if err := o.transitionTo(Preparing); err != nil {
		return err
	}

	return nil
}

func (o *Order) ready() error {
	if err := o.transitionTo(Ready); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigning() error {
	if err := o.transitionTo(Assigning); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigned() error {
	if err := o.transitionTo(Assigned); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivering() error {
	if err := o.transitionTo(Delivering); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivered() error {
	if err := o.transitionTo(Delivered); err != nil {
		return err
	}

	return nil
}

func (o *Order) failedAssign() (Compensation, error) {
	comp := compensations[transition{o.status, FailedAssign}]

	if err := o.transitionTo(FailedAssign); err != nil {
		return CompUnset, err
	}

	return comp, nil
}

// getters

func (o *Order) ID() uuid.UUID {
	return o.id
}

func (o *Order) UserID() uuid.UUID {
	return o.userID
}

func (o *Order) RestaurantID() uuid.UUID {
	return o.restaurantID
}

func (o *Order) TotalPrice() int64 {
	return o.totalPrice
}

func (o *Order) DeliveryAddress() string {
	return o.deliveryAddress
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) Status() Status {
	return o.status
}

func (o *Order) Items() []Item {
	return slices.Clone(o.items)
}
