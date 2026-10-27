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

func applyTransition(o *order.Order, to order.Status) (order.Compensation, *order.InvalidTransitionError) {
	switch to {
	case order.Paid:
		return order.CompUnset, order.PaidForTest(o)
	case order.Preparing:
		return order.CompUnset, order.PreparingForTest(o)
	case order.Ready:
		return order.CompUnset, order.ReadyForTest(o)
	case order.Assigning:
		return order.CompUnset, order.AssigningForTest(o)
	case order.Assigned:
		return order.CompUnset, order.AssignedForTest(o)
	case order.Delivering:
		return order.CompUnset, order.DeliveringForTest(o)
	case order.Delivered:
		return order.CompUnset, order.DeliveredForTest(o)
	case order.Cancelled:
		return order.CancelForTest(o)
	case order.FailedAssign:
		return order.FailedAssignForTest(o)
	default:
		panic("order: no transition method targets status" + string(to))
	}
}

func transitionTargets() []order.Status {
	return []order.Status{
		order.Paid, order.Preparing, order.Ready, order.Assigning,
		order.Assigned, order.Delivering, order.Delivered,
		order.Cancelled, order.FailedAssign,
	}
}

func TestTransitionToMatrix(t *testing.T) {
	t.Parallel()

	for _, from := range allStatuses() {
		for _, to := range transitionTargets() {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				t.Parallel()

				o := newOrderInStatus(t, from)

				comp, err := applyTransition(o, to)

				if allowedTransitions[from][to] {
					require.Nil(t, err)
					assert.Equal(t, to, o.Status(), "successful transition should change the status")
				} else {
					require.NotNil(t, err)
					assert.Equal(t, &order.InvalidTransitionError{From: from, To: to}, err,
						"error must contain both statuses")
					assert.Equal(t, from, o.Status(), "an unsuccessful transition should not change the status")

					if to == order.Cancelled || to == order.FailedAssign {
						assert.Equal(t, order.CompUnset, comp, "rejected transition must be unset")
					}
				}
			})
		}
	}
}

func TestInvalidTransitionErrorMessage(t *testing.T) {
	t.Parallel()

	err := &order.InvalidTransitionError{From: order.Created, To: order.Delivered}

	assert.EqualError(t, err, "order: invalid transition created -> delivered")
}

func TestTransitionToFullLifeCycle(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	for _, to := range []order.Status{
		order.Paid, order.Preparing, order.Ready, order.Assigning, order.Assigned, order.Delivering, order.Delivered,
	} {
		_, err := applyTransition(o, to)
		require.Nil(t, err, "transition to %s should be allowed", to)
	}

	assert.Equal(t, order.Delivered, o.Status())
}
