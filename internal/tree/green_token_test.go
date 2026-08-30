package tree

import (
	"testing"

	"github.com/arxbombus/jominia/internal/text"
)

func expectGreenTokenPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func TestGreenTokenWithoutTrivia(t *testing.T) {
	kind := RawSyntaxKind(1)
	token := NewGreenToken(kind, "foo")

	if token.Kind() != kind {
		t.Fatalf("Kind() = %v, want %v", token.Kind(), kind)
	}
	if token.Text() != "foo" {
		t.Fatalf("Text() = %q, want %q", token.Text(), "foo")
	}
	if token.TextTrimmed() != "foo" {
		t.Fatalf("TextTrimmed() = %q, want %q", token.TextTrimmed(), "foo")
	}
	if token.TextLen() != text.SizeOf("foo") {
		t.Fatalf("TextLen() = %d, want %d", token.TextLen(), text.SizeOf("foo"))
	}
	if token.LeadingTriviaTextLen() != 0 || token.TrailingTriviaTextLen() != 0 {
		t.Fatal("token without trivia reported non-zero trivia text length")
	}
	if token.LeadingTriviaCount() != 0 || token.TrailingTriviaCount() != 0 {
		t.Fatal("token without trivia reported trivia pieces")
	}
}

func TestGreenTokenWithTrivia(t *testing.T) {
	leading := []TriviaPiece{
		NewNewlineTriviaPiece(1),
		NewWhitespaceTriviaPiece(2),
	}
	trailing := []TriviaPiece{
		NewWhitespaceTriviaPiece(1),
		NewCommentTriviaPiece(4),
	}
	const value = "\n  foo # hi"

	token := NewGreenTokenWithTrivia(
		RawSyntaxKind(1),
		value,
		leading,
		trailing,
	)

	if token.Text() != value {
		t.Fatalf("Text() = %q, want %q", token.Text(), value)
	}
	if token.TextTrimmed() != "foo" {
		t.Fatalf("TextTrimmed() = %q, want %q", token.TextTrimmed(), "foo")
	}
	if token.TextLen() != text.SizeOf(value) {
		t.Fatalf("TextLen() = %d, want %d", token.TextLen(), text.SizeOf(value))
	}
	if token.LeadingTriviaTextLen() != 3 {
		t.Fatalf("LeadingTriviaTextLen() = %d, want 3", token.LeadingTriviaTextLen())
	}
	if token.TrailingTriviaTextLen() != 5 {
		t.Fatalf("TrailingTriviaTextLen() = %d, want 5", token.TrailingTriviaTextLen())
	}
	if token.LeadingTriviaCount() != 2 {
		t.Fatalf("LeadingTriviaCount() = %d, want 2", token.LeadingTriviaCount())
	}
	if token.TrailingTriviaCount() != 2 {
		t.Fatalf("TrailingTriviaCount() = %d, want 2", token.TrailingTriviaCount())
	}
	if token.LeadingTriviaPiece(0).Kind() != TriviaNewline {
		t.Fatalf("first leading trivia kind = %v, want %v", token.LeadingTriviaPiece(0).Kind(), TriviaNewline)
	}
	if token.TrailingTriviaPiece(1).Kind() != TriviaComment {
		t.Fatalf("second trailing trivia kind = %v, want %v", token.TrailingTriviaPiece(1).Kind(), TriviaComment)
	}
}

func TestGreenTokenCopiesTriviaSlices(t *testing.T) {
	leading := []TriviaPiece{NewWhitespaceTriviaPiece(1)}
	trailing := []TriviaPiece{NewCommentTriviaPiece(4)}
	token := NewGreenTokenWithTrivia(RawSyntaxKind(1), " foo#bar", leading, trailing)

	leading[0] = NewNewlineTriviaPiece(1)
	trailing[0] = NewWhitespaceTriviaPiece(4)

	if token.LeadingTriviaPiece(0).Kind() != TriviaWhitespace {
		t.Fatal("leading trivia changed after caller mutated the original slice")
	}
	if token.TrailingTriviaPiece(0).Kind() != TriviaComment {
		t.Fatal("trailing trivia changed after caller mutated the original slice")
	}
}

func TestGreenTokenRejectsTriviaLongerThanText(t *testing.T) {
	t.Run("leading", func(t *testing.T) {
		expectGreenTokenPanic(t, func() {
			_ = NewGreenTokenWithTrivia(
				RawSyntaxKind(1),
				"foo",
				[]TriviaPiece{NewWhitespaceTriviaPiece(4)},
				nil,
			)
		})
	})

	t.Run("combined", func(t *testing.T) {
		expectGreenTokenPanic(t, func() {
			_ = NewGreenTokenWithTrivia(
				RawSyntaxKind(1),
				"foo",
				[]TriviaPiece{NewWhitespaceTriviaPiece(2)},
				[]TriviaPiece{NewWhitespaceTriviaPiece(2)},
			)
		})
	})
}
