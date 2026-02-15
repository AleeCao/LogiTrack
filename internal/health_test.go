package internal

import "testing"

func TestSanity(t *testing.T) {
	want := 1
	got := 1
	if got != want {
		t.Errorf("Math is broken: got %d, want %d", got, want)
	}
}
