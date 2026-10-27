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
	ErrInvalidStatus          = errors.New("order: restore with invalid status")
	ErrInvalidID              = errors.New("order: id is nil")
)

const maxDeliveryAdressLen = 300

type Item struct {
	DishID    uuid.UUID
	Name      string
	UnitPrice int64
	Quantity  int
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

func (o *Order) cancel() (Compensation, *InvalidTransitionError) {
	comp := compensations[o.status]
	if err := o.transitionTo(Cancelled); err != nil {
		return CompUnset, err
	}

	return comp, nil
}

func (o *Order) paid() *InvalidTransitionError {
	if err := o.transitionTo(Paid); err != nil {
		return err
	}

	return nil
}

func (o *Order) preparing() *InvalidTransitionError {
	if err := o.transitionTo(Preparing); err != nil {
		return err
	}

	return nil
}

func (o *Order) ready() *InvalidTransitionError {
	if err := o.transitionTo(Ready); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigning() *InvalidTransitionError {
	if err := o.transitionTo(Assigning); err != nil {
		return err
	}

	return nil
}

func (o *Order) assigned() *InvalidTransitionError {
	if err := o.transitionTo(Assigned); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivering() *InvalidTransitionError {
	if err := o.transitionTo(Delivering); err != nil {
		return err
	}

	return nil
}

func (o *Order) delivered() *InvalidTransitionError {
	if err := o.transitionTo(Delivered); err != nil {
		return err
	}

	return nil
}

func (o *Order) failedAssign() (Compensation, *InvalidTransitionError) {
	comp := compensations[o.status]

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
