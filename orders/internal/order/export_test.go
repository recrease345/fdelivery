package order

func CancelForTest(o *Order) (Compensation, error) {
	return o.cancel()
}

func PaidForTest(o *Order) error {
	return o.paid()
}

func PreparingForTest(o *Order) error {
	return o.preparing()
}

func ReadyForTest(o *Order) error {
	return o.ready()
}

func AssigningForTest(o *Order) error {
	return o.assigning()
}

func AssignedForTest(o *Order) error {
	return o.assigned()
}

func DeliveringForTest(o *Order) error {
	return o.delivering()
}

func DeliveredForTest(o *Order) error {
	return o.delivered()
}

func FailedAssignForTest(o *Order) (Compensation, error) {
	return o.failedAssign()
}
