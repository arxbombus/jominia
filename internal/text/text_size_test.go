package text

import "testing"

func TestSizeOf(t *testing.T) {
	if got := SizeOf("hello"); got != 5 {
		t.Fatalf("SizeOf(\"hello\") = %d, want 5", got)
	}

	if got := SizeOf("🦀"); got != 4 {
		t.Fatalf("SizeOf(\"🦀\") = %d, want 4", got)
	}
}
