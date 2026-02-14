package passing

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Errorf("Add(2, 3) = %d, want 5", got)
	}
}

func TestMul(t *testing.T) {
	if got := Mul(4, 5); got != 20 {
		t.Errorf("Mul(4, 5) = %d, want 20", got)
	}
}
