package order_test

import (
	"fdelivery_orders/internal/order"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreValidation(t *testing.T) {
	t.Parallel()

	snapshot := newOrder(t)

	tests := []struct {
		name    string
		id      uuid.UUID
		status  order.Status
		wantErr error
	}{
		{"nil id", uuid.Nil, order.Created, order.ErrInvalidID},
		{"unknown status", uuid.New(), order.Status("unknown"), order.ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := order.Restore(tt.id, snapshot.UserID(), snapshot.RestaurantID(),
				snapshot.TotalPrice(), snapshot.Items(), snapshot.DeliveryAddress(), snapshot.CreatedAt(), tt.status)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got, "on validation error restore must retunr nil")
		})
	}
}

func TestRestoreRoundTrip(t *testing.T) {
	t.Parallel()

	src := newOrder(t)

	got, err := order.Restore(
		src.ID(), src.UserID(), src.RestaurantID(), src.TotalPrice(), src.Items(), src.DeliveryAddress(), src.CreatedAt(), src.Status(),
	)

	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, src.ID(), got.ID())
	assert.Equal(t, src.UserID(), got.UserID())
	assert.Equal(t, src.RestaurantID(), got.RestaurantID())
	assert.Equal(t, src.TotalPrice(), got.TotalPrice())
	assert.Equal(t, src.Items(), got.Items())
	assert.Equal(t, src.DeliveryAddress(), got.DeliveryAddress())
	assert.Equal(t, src.CreatedAt(), got.CreatedAt())
	assert.Equal(t, src.Status(), got.Status())
}

func TestRestoreEveryStatus(t *testing.T) {
	t.Parallel()

	snapshot := newOrder(t)

	for _, s := range allStatuses() {
		t.Run(string(s), func(t *testing.T) {
			t.Parallel()

			got, err := order.Restore(snapshot.ID(), snapshot.UserID(), snapshot.RestaurantID(),
				snapshot.TotalPrice(), snapshot.Items(), snapshot.DeliveryAddress(),
				snapshot.CreatedAt(), s)

			require.NoError(t, err, "every known status must be restorable")
			assert.Equal(t, s, got.Status())
		})
	}
}
