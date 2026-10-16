package order

func CancelForTest(o *Order) (Compensation, *InvalidTransitionError) {
	return o.cancel()
}

func PaidForTest(o *Order) *InvalidTransitionError {
	return o.paid()
}

func PreparingForTest(o *Order) *InvalidTransitionError {
	return o.preparing()
}

func ReadyForTest(o *Order) *InvalidTransitionError {
	return o.ready()
}

func AssigningForTest(o *Order) *InvalidTransitionError {
	return o.assigning()
}

func AssignedForTest(o *Order) *InvalidTransitionError {
	return o.assigned()
}

func DeliveringForTest(o *Order) *InvalidTransitionError {
	return o.delivering()
}

func DeliveredForTest(o *Order) *InvalidTransitionError {
	return o.delivered()
}

func FailedAssignForTest(o *Order) (Compensation, *InvalidTransitionError) {
	return o.failedAssign()
}
