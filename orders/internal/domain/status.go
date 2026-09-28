package order

import (
	"fmt"
)

type Status string

const (
	Created      Status = "created"
	Paid         Status = "paid"
	Preparing    Status = "preparing"
	Ready        Status = "ready"
	Assigning    Status = "assigning"
	Assigned     Status = "assigned"
	Delivering   Status = "delivering"
	Delivered    Status = "delivered"
	Cancelled    Status = "cancelled"
	FailedAssign Status = "failed_assign"
)

type InvalidTransitionError struct {
	From, To Status
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("order: invalid transition %s -> %s", e.From, e.To)
}

var transitions = map[Status][]Status{
	Created:    {Paid, Cancelled},
	Paid:       {Preparing, Cancelled},
	Preparing:  {Ready, Cancelled},
	Ready:      {Assigning, Cancelled},
	Assigning:  {Assigned, FailedAssign, Cancelled},
	Assigned:   {Delivering, Cancelled},
	Delivering: {Delivered},
}

type Compensation int

const (
	CompUnset Compensation = 0
	CompNone  Compensation = iota + 1
	CompVoid
	CompVoidRestaurant
	CompVoidReleaseDriverRestaurant
	CompForbidden
)

var compensations = map[Status]Compensation{
	Created:    CompNone,
	Paid:       CompVoid,
	Preparing:  CompVoidRestaurant,
	Ready:      CompVoidRestaurant,
	Assigning:  CompVoidReleaseDriverRestaurant,
	Assigned:   CompVoidReleaseDriverRestaurant,
	Delivering: CompForbidden,
}

func (order *Order) TransitionTo(to Status) *InvalidTransitionError {
	for _, status := range transitions[order.Status] {
		if status == to {
			order.Status = to
			return nil
		}
	}

	return &InvalidTransitionError{
		From: order.Status,
		To:   to,
	}
}
