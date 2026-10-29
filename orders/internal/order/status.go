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
	Assigned:   {Delivering, Cancelled, Assigning},
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

type transition struct{ From, To Status }

var compensations = map[transition]Compensation{
	{Created, Cancelled}:      CompNone,
	{Paid, Cancelled}:         CompVoid,
	{Preparing, Cancelled}:    CompVoidRestaurant,
	{Ready, Cancelled}:        CompVoidRestaurant,
	{Assigning, Cancelled}:    CompVoidReleaseDriverRestaurant,
	{Assigning, FailedAssign}: CompVoid,
	{Assigned, Cancelled}:     CompVoidReleaseDriverRestaurant,
	{Assigned, Assigning}:     CompNone,
}

func (order *Order) transitionTo(to Status) *InvalidTransitionError {
	if !CanTransitionTo(order.Status(), to) {
		return &InvalidTransitionError{From: order.status, To: to}
	}

	order.status = to
	return nil
}

func CanTransitionTo(from, to Status) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}

	return false
}
