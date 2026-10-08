package order_test

import (
	"fdelivery_orders/internal/order"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allowedTransitions = map[order.Status]map[order.Status]bool{
	order.Created:    {order.Paid: true, order.Cancelled: true},
	order.Paid:       {order.Preparing: true, order.Cancelled: true},
	order.Preparing:  {order.Ready: true, order.Cancelled: true},
	order.Ready:      {order.Assigning: true, order.Cancelled: true},
	order.Assigning:  {order.Assigned: true, order.FailedAssign: true, order.Cancelled: true},
	order.Assigned:   {order.Delivering: true, order.Cancelled: true},
	order.Delivering: {order.Delivered: true},
}

func TestTransitionTo_Matrix(t *testing.T) {
	t.Parallel()

	for _, from := range allStatuses() {
		for _, to := range allStatuses() {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				t.Parallel()

				o := newOrderInStatus(t, from)

				err := o.TransitionTo(to)

				if allowedTransitions[from][to] {
					require.Nil(t, err)
					assert.Equal(t, to, o.Status, "successful transition should change the status")
				} else {
					require.NotNil(t, err)
					assert.Equal(t, &order.InvalidTransitionError{From: from, To: to}, err,
						"error must contain both statuses")
					assert.Equal(t, from, o.Status, "an unsuccessful transition should not change the status")
				}
			})
		}
	}
}

func TestTransitionToFullLifeCycle(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	for _, s := range []order.Status{
		order.Paid, order.Preparing, order.Ready, order.Assigning, order.Assigned, order.Delivering, order.Delivered,
	} {
		require.Nil(t, o.TransitionTo(s), "transition to %s should be allowed", s)
	}

	assert.Equal(t, order.Delivered, o.Status)
}

func TestTransitionToUnknownTargetStatus(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	err := o.TransitionTo(order.Status("unknownStatus"))

	require.NotNil(t, err)

	assert.Equal(t, &order.InvalidTransitionError{From: order.Created, To: order.Status("unknownStatus")}, err)
}
