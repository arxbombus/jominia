package tree

import (
	"testing"

	"github.com/arxbombus/jominia/internal/text"
)

func expectTriviaPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func TestTriviaPieceKindPredicates(t *testing.T) {
	if !TriviaNewline.IsNewline() || TriviaNewline.IsWhitespace() || TriviaNewline.IsComment() {
		t.Fatal("TriviaNewline predicates are incorrect")
	}
	if !TriviaWhitespace.IsWhitespace() || TriviaWhitespace.IsNewline() || TriviaWhitespace.IsComment() {
		t.Fatal("TriviaWhitespace predicates are incorrect")
	}
	if !TriviaComment.IsComment() || TriviaComment.IsNewline() || TriviaComment.IsWhitespace() {
		t.Fatal("TriviaComment predicates are incorrect")
	}
}

func TestTriviaPieceConstructors(t *testing.T) {
	tests := []struct {
		name   string
		piece  TriviaPiece
		kind   TriviaPieceKind
		length text.TextSize
	}{
		{"generic", NewTriviaPiece(TriviaWhitespace, 3), TriviaWhitespace, 3},
		{"newline", NewNewlineTriviaPiece(2), TriviaNewline, 2},
		{"whitespace", NewWhitespaceTriviaPiece(4), TriviaWhitespace, 4},
		{"comment", NewCommentTriviaPiece(7), TriviaComment, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.piece.Kind() != tt.kind {
				t.Fatalf("Kind() = %v, want %v", tt.piece.Kind(), tt.kind)
			}
			if tt.piece.TextLen() != tt.length {
				t.Fatalf("TextLen() = %d, want %d", tt.piece.TextLen(), tt.length)
			}
		})
	}
}

func TestNewTriviaPieceRejectsInvalidKind(t *testing.T) {
	expectTriviaPanic(t, func() {
		_ = NewTriviaPiece(TriviaPieceKind(255), 1)
	})
}

func TestGreenTriviaEmptyZeroValue(t *testing.T) {
	var trivia greenTrivia

	if trivia.data != nil {
		t.Fatal("zero-value greenTrivia should have nil data")
	}
	if trivia.len() != 0 {
		t.Fatalf("len() = %d, want 0", trivia.len())
	}
	if trivia.textLen() != 0 {
		t.Fatalf("textLen() = %d, want 0", trivia.textLen())
	}
}

func TestGreenTriviaCopiesPiecesAndCalculatesLength(t *testing.T) {
	pieces := []TriviaPiece{
		NewWhitespaceTriviaPiece(2),
		NewNewlineTriviaPiece(1),
		NewCommentTriviaPiece(5),
	}
	trivia := newGreenTrivia(pieces)

	if trivia.len() != 3 {
		t.Fatalf("len() = %d, want 3", trivia.len())
	}
	if trivia.textLen() != 8 {
		t.Fatalf("textLen() = %d, want 8", trivia.textLen())
	}

	pieces[0] = NewCommentTriviaPiece(99)
	if trivia.piece(0).Kind() != TriviaWhitespace || trivia.piece(0).TextLen() != 2 {
		t.Fatal("green trivia changed after caller mutated the original slice")
	}
}

func TestGreenTriviaHandleCopiesShareData(t *testing.T) {
	first := newGreenTrivia([]TriviaPiece{NewWhitespaceTriviaPiece(2)})
	second := first

	if first.data == nil {
		t.Fatal("expected non-empty green trivia data")
	}
	if first.data != second.data {
		t.Fatal("copying greenTrivia should copy the handle and share immutable data")
	}
}

func TestGreenTriviaPieceOnEmptyPanics(t *testing.T) {
	var trivia greenTrivia

	expectTriviaPanic(t, func() {
		_ = trivia.piece(0)
	})
}

func TestGreenTriviaRejectsTextLengthOverflow(t *testing.T) {
	pieces := []TriviaPiece{
		NewWhitespaceTriviaPiece(^text.TextSize(0)),
		NewWhitespaceTriviaPiece(1),
	}

	expectTriviaPanic(t, func() {
		_ = newGreenTrivia(pieces)
	})
}
