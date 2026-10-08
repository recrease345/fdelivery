package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompensationFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status Status
		want   Compensation
	}{
		{"created", Created, CompNone},
		{"paid", Paid, CompVoid},
		{"preparing", Preparing, CompVoidRestaurant},
		{"ready", Ready, CompVoidRestaurant},
		{"assigning", Assigning, CompVoidReleaseDriverRestaurant},
		{"assigned", Assigned, CompVoidReleaseDriverRestaurant},
		{"delivering", Delivering, CompForbidden},
		{"terminal cancelled", Cancelled, CompUnset},
		{"terminal delivered", Delivered, CompUnset},
		{"terminal failed assign", FailedAssign, CompUnset},
		{"unknown status", Status("unknownStatus"), CompUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, CompensationFor(tt.status))
		})
	}
}

func TestEveryCancellableStatusHasCompensation(t *testing.T) {
	t.Parallel()

	for from, targets := range transitions {
		for _, to := range targets {
			if to == Cancelled && compensations[from] == CompUnset {
				t.Errorf("status %q is cancellable but compensation policy is not defined", from)
			}
		}
	}
}
