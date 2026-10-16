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
	CompUnset Compensation = iota
	CompNone  
	CompVoid
	CompVoidRestaurant
	CompVoidReleaseDriverRestaurant
)

var compensations = map[Status]Compensation{
	Created:   CompNone,
	Paid:      CompVoid,
	Preparing: CompVoidRestaurant,
	Ready:     CompVoidRestaurant,
	Assigning: CompVoidReleaseDriverRestaurant,
	Assigned:  CompVoidReleaseDriverRestaurant,
}

func (order *Order) transitionTo(to Status) *InvalidTransitionError {
	for _, status := range transitions[order.status] {
		if status == to {
			order.status = to
			return nil
		}
	}

	return &InvalidTransitionError{
		From: order.status,
		To:   to,
	}
}
