package order_test

import (
	order "fdelivery_orders/internal/domain"
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

func totalOf(items []order.Item) int64 {
	var sum int64
	for _, item := range items {
		sum += item.UnitPrice * int64(item.Quantity)
	}

	return sum
}

// изначально валидна, потом по одному полю ломается
type fixture struct {
	userID          uuid.UUID
	restaurantID    uuid.UUID
	items           []order.Item
	totalPrice      int64
	deliveryAddress string
	now             time.Time
}

func validFixture() fixture {
	items := []order.Item{validItem(), anotherItem()}

	return fixture{
		userID:          uuid.New(),
		restaurantID:    uuid.New(),
		items:           items,
		totalPrice:      totalOf(items),
		deliveryAddress: "Улица Пушкина Дом колотушкина",
		now:             fixedNow,
	}
}

func newOrder(t testing.TB) *order.Order {
	t.Helper()

	f := validFixture()
	order, err := order.NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, f.now)
	require.NoError(t, err, "fixture: valid data must not cause an error")
	return order
}

func newOrderInStatus(t testing.TB, s order.Status) *order.Order {
	t.Helper()

	order := newOrder(t)
	order.Status = s
	return order
}

func allStatuses() []order.Status {
	return []order.Status{
		order.Created, order.Paid, order.Preparing, order.Ready, order.Assigning, order.Assigned, order.Delivering, order.Delivered, order.Cancelled, order.FailedAssign,
	}
}
