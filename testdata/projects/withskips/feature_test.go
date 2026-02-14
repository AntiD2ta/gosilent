package withskips

import "testing"

func TestIsReady(t *testing.T) {
	if !IsReady() {
		t.Error("expected IsReady() to return true")
	}
}

func TestFuture(t *testing.T) {
	t.Skip("not implemented yet")
}

func TestExperimental(t *testing.T) {
	t.Skip("requires external service")
}
