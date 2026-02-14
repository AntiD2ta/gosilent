package failing

import "testing"

func TestDouble(t *testing.T) {
	if got := Double(3); got != 6 {
		t.Errorf("Double(3) = %d, want 6", got)
	}
}

func TestDoubleBroken(t *testing.T) {
	// Deliberately wrong expectation to force a failure.
	if got := Double(5); got != 99 {
		t.Errorf("Double(5) = %d, want 99", got)
	}
}
