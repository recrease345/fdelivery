package order_test

import (
	"fdelivery_orders/internal/order"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryCancellableStatusHasCompensation(t *testing.T) {
	t.Parallel()

	for _, from := range allStatuses() {
		t.Run(string(from), func(t *testing.T) {
			t.Parallel()

			o := newOrderInStatus(t, from)

			comp, err := order.CancelForTest(o)
			if err == nil {
				assert.NotEqual(t, order.CompUnset, comp, "cancel from %s succeded, compensation must be defined", from)
				return
			}

			assert.Equal(t, order.CompUnset, comp, "cancel from %s rejected, it must not have a compensation")
		})
	}
}

func TestCancelCompensationContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		from     order.Status
		wantComp order.Compensation
	}{
		{"cancel right after creation", order.Created, order.CompNone},
		{"cancel after payment, void the charge", order.Paid, order.CompVoid},
		{"cancel while the kitchen is cooking", order.Preparing, order.CompVoidRestaurant},
		{"cancel when food is ready", order.Ready, order.CompVoidRestaurant},
		{"cancel while searching for a courier", order.Assigning, order.CompVoidReleaseDriverRestaurant},
		{"cancel after courier was assigned", order.Assigned, order.CompVoidReleaseDriverRestaurant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			o := newOrderInStatus(t, tt.from)

			comp, err := order.CancelForTest(o)

			require.Nil(t, err, "cancel from %s must be allowed", tt.from)
			assert.Equal(t, order.Cancelled, o.Status())
			assert.Equal(t, tt.wantComp, comp, "compensation must reflect status we cancel from")
		})
	}
}

func TestFailedAssignCompensation(t *testing.T) {
	t.Parallel()

	o := newOrderInStatus(t, order.Assigning)

	comp, err := order.FailedAssignForTest(o)

	require.Nil(t, err)
	assert.Equal(t, order.FailedAssign, o.Status())
	assert.Equal(t, order.CompVoidReleaseDriverRestaurant, comp)
}
