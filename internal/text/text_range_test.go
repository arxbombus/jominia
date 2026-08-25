package text

import "testing"

func TestRangeLength(t *testing.T) {
	r := NewTextRange(10, 15)

	if got := r.Len(); got != 5 {
		t.Fatalf("length = %d, want 5", got)
	}
}

func TestRangeStartAndEnd(t *testing.T) {
	r := NewTextRange(10, 15)

	if got := r.Start(); got != 10 {
		t.Fatalf("start = %d, want 10", got)
	}

	if got := r.End(); got != 15 {
		t.Fatalf("end = %d, want 15", got)
	}
}

func TestEmptyRange(t *testing.T) {
	r := NewTextRange(10, 10)

	if !r.IsEmpty() {
		t.Fatal("expected range to be empty")
	}
}

func TestRangeContains(t *testing.T) {
	r := NewTextRange(10, 15)

	if !r.Contains(10) {
		t.Error("expected range to contain its start")
	}

	if !r.Contains(14) {
		t.Error("expected range to contain offset 14")
	}

	if r.Contains(15) {
		t.Error("expected range not to contain its end")
	}
}
