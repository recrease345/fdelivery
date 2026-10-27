package order

import "testing"

func TestEveryCompensableTransitionHasCompensation(t *testing.T) {
	t.Parallel()

	for from, targets := range transitions {
		for _, to := range targets {
			if to != Cancelled && to != FailedAssign {
				continue
			}
			if compensations[transition{from, to}] == CompUnset {
				t.Errorf("transition %q -> %q is allowed, but compensation is unset", from, to)
			}
		}
	}
}

func TestEveryCompensationHasAllowedTransition(t *testing.T) {
	t.Parallel()

	for tr := range compensations {
		allowed := false
		for _, to := range transitions[tr.From] {
			if to == tr.To {
				allowed = true
				break
			}
		}
		if !allowed {
			t.Errorf("compensation for %q -> %q is defined, but the transition does not exist", tr.From, tr.To)
		}
	}
}
