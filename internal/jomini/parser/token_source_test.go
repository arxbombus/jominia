package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
)

func TestTokenSourceSkipsTrivia(t *testing.T) {
	ts := NewTokenSource("foo   = bar")

	if got := ts.Current(); got != syntax.Identifier {
		t.Fatalf("expected Identifier, got %s", got)
	}

	ts.Bump()

	if got := ts.Current(); got != syntax.Equals {
		t.Fatalf("expected Equals, got %s", got)
	}

	ts.Bump()

	if got := ts.Current(); got != syntax.Identifier {
		t.Fatalf("expected Identifier, got %s", got)
	}

	ts.Bump()

	if got := ts.Current(); got != syntax.EOF {
		t.Fatalf("expected EOF, got %s", got)
	}
}

func TestTokenSourceCurrentRange(t *testing.T) {
	ts := NewTokenSource("foo = bar")

	r := ts.CurrentRange()

	if r.Start() != 0 {
		t.Errorf("expected start 0, got %d", r.Start())
	}

	if r.End() != 3 {
		t.Errorf("expected end 3, got %d", r.End())
	}

	ts.Bump()

	r = ts.CurrentRange()

	if r.Start() != 4 {
		t.Errorf("expected start 4, got %d", r.Start())
	}

	if r.End() != 5 {
		t.Errorf("expected end 5, got %d", r.End())
	}
}

func TestTokenSourcePrecedingLineBreak(t *testing.T) {
	ts := NewTokenSource("foo bar\nbaz")

	if ts.HasPrecedingLineBreak() {
		t.Fatal("foo should not have a preceding line break")
	}

	ts.Bump()

	if ts.HasPrecedingLineBreak() {
		t.Fatal("bar should not have a preceding line break")
	}

	ts.Bump()

	if !ts.HasPrecedingLineBreak() {
		t.Fatal("baz should have a preceding line break")
	}
}

func TestTokenSourceTrivia(t *testing.T) {
	ts := NewTokenSource("foo   # hello\n    baz")

	// Move from foo to baz.
	ts.Bump()

	trivia := ts.Finish()

	if len(trivia) != 4 {
		t.Fatalf("expected 4 trivia pieces, got %d", len(trivia))
	}

	if trivia[0].Kind != syntax.Whitespace || !trivia[0].IsTrailing {
		t.Errorf("expected first trivia to be trailing whitespace")
	}

	if trivia[1].Kind != syntax.Comment || !trivia[1].IsTrailing {
		t.Errorf("expected second trivia to be trailing comment")
	}

	if trivia[2].Kind != syntax.Newline || trivia[2].IsTrailing {
		t.Errorf("expected newline to be leading trivia")
	}

	if trivia[3].Kind != syntax.Whitespace || trivia[3].IsTrailing {
		t.Errorf("expected indentation to be leading trivia")
	}
}

func TestTokenSourceInitialTriviaIsLeading(t *testing.T) {
	ts := NewTokenSource("   # hello\nfoo")

	trivia := ts.Finish()

	if len(trivia) != 3 {
		t.Fatalf("expected 3 trivia pieces, got %d", len(trivia))
	}

	for i, piece := range trivia {
		if piece.IsTrailing {
			t.Errorf("trivia %d should be leading", i)
		}
	}
}

func TestTokenSourceFinalTrivia(t *testing.T) {
	ts := NewTokenSource("foo\n# goodbye\n")

	ts.Bump()

	if ts.Current() != syntax.EOF {
		t.Fatalf("expected EOF, got %s", ts.Current())
	}

	trivia := ts.Finish()

	if len(trivia) != 3 {
		t.Fatalf("expected 3 trivia pieces, got %d", len(trivia))
	}

	for i, piece := range trivia {
		if piece.IsTrailing {
			t.Errorf("trivia %d should be leading", i)
		}
	}
}
