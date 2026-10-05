package order_test

import (
	order "fdelivery_orders/internal/domain/order"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompensationFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status order.Status
		want   order.Compensation
	}{
		{"created", order.Created, order.CompNone},
		{"paid", order.Paid, order.CompVoid},
		{"preparing", order.Preparing, order.CompVoidRestaurant},
		{"ready", order.Ready, order.CompVoidRestaurant},
		{"assigning", order.Assigning, order.CompVoidReleaseDriverRestaurant},
		{"assigned", order.Assigned, order.CompVoidReleaseDriverRestaurant},
		{"delivering", order.Delivering, order.CompForbidden},
		{"terminal cancelled", order.Cancelled, order.CompUnset},
		{"terminal delivered", order.Delivered, order.CompUnset},
		{"terminal failed assign", order.FailedAssign, order.CompUnset},
		{"unknown status", order.Status("unknownStatus"), order.CompUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, order.CompensationFor(tt.status))
		})
	}
}

func TestCompensationDefinedForEveryCancellableStatus(t *testing.T) {
	t.Parallel()

	for from, targets := range allowedTransitions {
		if !targets[order.Cancelled] {
			continue
		}
		if order.CompensationFor(from) == order.CompUnset {
			t.Errorf("status %s can be changed but compensation policy has not been defined", from)
		}
	}
}
