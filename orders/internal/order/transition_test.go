package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allowedTransitions = map[Status]map[Status]bool{
	Created:    {Paid: true, Cancelled: true},
	Paid:       {Preparing: true, Cancelled: true},
	Preparing:  {Ready: true, Cancelled: true},
	Ready:      {Assigning: true, Cancelled: true},
	Assigning:  {Assigned: true, FailedAssign: true, Cancelled: true},
	Assigned:   {Delivering: true, Cancelled: true},
	Delivering: {Delivered: true},
}

func TestTransitionTo_Matrix(t *testing.T) {
	t.Parallel()

	for _, from := range allStatuses() {
		for _, to := range allStatuses() {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				t.Parallel()

				o := newOrderInStatus(t, from)

				err := o.transitionTo(to)

				if allowedTransitions[from][to] {
					require.Nil(t, err)
					assert.Equal(t, to, o.Status, "successful transition should change the status")
				} else {
					require.NotNil(t, err)
					assert.Equal(t, &InvalidTransitionError{From: from, To: to}, err,
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

	for _, s := range []Status{
		Paid, Preparing, Ready, Assigning, Assigned, Delivering, Delivered,
	} {
		require.Nil(t, o.transitionTo(s), "transition to %s should be allowed", s)
	}

	assert.Equal(t, Delivered, o.Status)
}

func TestTransitionToUnknownTargetStatus(t *testing.T) {
	t.Parallel()

	o := newOrder(t)

	err := o.transitionTo(Status("unknownStatus"))

	require.NotNil(t, err)

	assert.Equal(t, &InvalidTransitionError{From: Created, To: Status("unknownStatus")}, err)
}
