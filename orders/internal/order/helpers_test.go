package order_test

import (
	"fdelivery_orders/internal/order"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

func validItem() order.Item {
	return order.Item{
		DishID:    uuid.New(),
		Name:      "пицца",
		UnitPrice: 59900,
		Quantity:  3,
	}
}

func anotherItem() order.Item {
	return order.Item{
		DishID:    uuid.New(),
		Name:      "тройной чизбургер дабл трипл воппер",
		UnitPrice: 29900,
		Quantity:  1,
	}
}

// изначально валидна, потом по одному полю ломается
type fixture struct {
	userID          uuid.UUID
	restaurantID    uuid.UUID
	items           []order.Item
	deliveryAddress string
	now             time.Time
}

func validFixture() fixture {
	items := []order.Item{validItem(), anotherItem()}

	return fixture{
		userID:          uuid.New(),
		restaurantID:    uuid.New(),
		items:           items,
		deliveryAddress: "Улица Пушкина Дом колотушкина",
		now:             fixedNow,
	}
}

func newOrder(t testing.TB) *order.Order {
	t.Helper()

	f := validFixture()
	o, err := order.NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, f.now)
	require.NoError(t, err, "fixture: valid data must not cause an error")
	return o
}

func newOrderInStatus(t testing.TB, s order.Status) *order.Order {
	t.Helper()

	o := newOrder(t)
	o.Status = s
	return o
}

func allStatuses() []order.Status {
	return []order.Status{
		order.Created, order.Paid, order.Preparing, order.Ready, order.Assigning, order.Assigned, order.Delivering, order.Delivered, order.Cancelled, order.FailedAssign,
	}
}
