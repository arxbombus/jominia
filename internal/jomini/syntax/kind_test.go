package syntax

import "testing"

func TestTombstoneIsZeroValue(t *testing.T) {
	var kind SyntaxKind

	if kind != Tombstone {
		t.Fatalf("zero value of SyntaxKind = %d, want Tombstone (%d)", kind, Tombstone)
	}
}

func TestSyntaxKindRawRoundTrip(t *testing.T) {
	kind := Identifier

	if got := FromRaw(kind.Raw()); got != kind {
		t.Fatalf("round trip = %d, want %d", got, kind)
	}
}

func TestTriviaKinds(t *testing.T) {
	if !Whitespace.IsTrivia() {
		t.Error("Whitespace should be trivia")
	}

	if !Newline.IsTrivia() {
		t.Error("Newline should be trivia")
	}

	if !Comment.IsTrivia() {
		t.Error("Comment should be trivia")
	}

	if Identifier.IsTrivia() {
		t.Error("Identifier should not be trivia")
	}
}
