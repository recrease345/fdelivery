package order_test

import (
	order "fdelivery_orders/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderSuccess(t *testing.T) {
	t.Parallel()

	f := validFixture()

	got, err := order.NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, f.userID, got.UserID)
	assert.Equal(t, f.restaurantID, got.RestaurantID)
	assert.Equal(t, f.items, got.Items)
	assert.Equal(t, f.deliveryAddress, got.DeliveryAddress)
	assert.Equal(t, f.deliveryAddress, got.DeliveryAddress)
	assert.Equal(t, f.deliveryAddress, got.DeliveryAddress)
	assert.Equal(t, f.now, got.CreatedAt)
	assert.Equal(t, order.Created, got.Status)
	assert.Equal(t, uuid.Nil, got.ID)
}

func TestNewOrderValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(f *fixture)
		wantErr error
	}{
		{"nil user id", func(f *fixture) { f.userID = uuid.Nil }, order.ErrInvalidUserID},
		{"nil restaraunt id", func(f *fixture) { f.restaurantID = uuid.Nil }, order.ErrInvalidRestaurantID},
		{"nil items", func(f *fixture) { f.items = nil }, order.ErrInvalidItems},
		{"empty  items", func(f *fixture) { f.items = []order.Item{} }, order.ErrInvalidItems},
		{"empty delivery address", func(f *fixture) { f.deliveryAddress = "" }, order.ErrInvalidDeliveryAddress},
		{"item empty name", func(f *fixture) { f.items[0].Name = "" }, order.ErrInvalidItemsAttributes},
		{"item nil dish id", func(f *fixture) { f.items[0].DishID = uuid.Nil }, order.ErrInvalidItemsAttributes},
		{"item zero quantity", func(f *fixture) { f.items[0].Quantity = 0 }, order.ErrInvalidItemsAttributes},
		{"item negative quantity", func(f *fixture) { f.items[0].Quantity = -1 }, order.ErrInvalidItemsAttributes},
		{"item zero unit price", func(f *fixture) { f.items[0].UnitPrice = 0 }, order.ErrInvalidItemsAttributes},
		{"item negative unit price", func(f *fixture) { f.items[0].UnitPrice = -1 }, order.ErrInvalidItemsAttributes},
		{"second item invalid", func(f *fixture) { f.items[1].Quantity = 0 }, order.ErrInvalidItemsAttributes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := validFixture()
			tt.mutate(&f)

			got, err := order.NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got, "constructor returned nil")
		})
	}
}

func TestNewOrderItemsAreDefensiveCopy(t *testing.T) {
	t.Parallel()

	f := validFixture()

	o, err := order.NewOrder(f.userID, f.restaurantID, f.items, f.deliveryAddress, fixedNow)
	require.NoError(t, err)

	f.items[0].Quantity = -100

	assert.Equal(t, 2, o.Items[0].Quantity, "mutation of the callers slice should not change it")
}
