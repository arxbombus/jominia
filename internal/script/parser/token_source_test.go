package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
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

func TestTokenSourceNthRange(t *testing.T) {
	ts := NewTokenSource("[[condition]")

	if got := ts.NthRange(0); got.Start() != 0 || got.End() != 1 {
		t.Fatalf("NthRange(0) = [%d,%d), want [0,1)", got.Start(), got.End())
	}
	if got := ts.NthRange(1); got.Start() != 1 || got.End() != 2 {
		t.Fatalf("NthRange(1) = [%d,%d), want [1,2)", got.Start(), got.End())
	}
}

func TestTokenSourceReLexesInlineMathAndRestoresNormalMode(t *testing.T) {
	ts := NewTokenSource(`@[1/3] next-value`)
	if ts.Current() != syntax.InlineMathStart {
		t.Fatalf("current = %s, want InlineMathStart", ts.Current())
	}

	ts.Bump()
	if ts.Current() != syntax.Identifier {
		t.Fatalf("normal interior = %s, want Identifier", ts.Current())
	}
	ts.ReLex(lexer.ReLexInlineMath)

	want := []syntax.SyntaxKind{
		syntax.Number,
		syntax.Slash,
		syntax.Number,
		syntax.RBracket,
		syntax.Identifier,
	}
	for _, kind := range want {
		if ts.Current() != kind {
			t.Fatalf("current = %s, want %s", ts.Current(), kind)
		}
		ts.Bump()
	}
}

func TestTokenSourceText(t *testing.T) {
	const source = "foo = bar"
	ts := NewTokenSource(source)

	if got := ts.Text(); got != source {
		t.Fatalf("expected %q, got %q", source, got)
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

func TestTokenSourceLookaheadDoesNotAdvance(t *testing.T) {
	ts := NewTokenSource("foo # comment\n= bar")

	if got := ts.Nth(0); got != syntax.Identifier {
		t.Fatalf("expected Nth(0) to be Identifier, got %s", got)
	}

	if got := ts.Nth(1); got != syntax.Equals {
		t.Fatalf("expected Nth(1) to be Equals, got %s", got)
	}

	if got := ts.Nth(2); got != syntax.Identifier {
		t.Fatalf("expected Nth(2) to be Identifier, got %s", got)
	}

	if got := ts.Current(); got != syntax.Identifier {
		t.Fatalf("lookahead advanced token source to %s", got)
	}

	ts.Bump()

	if got := ts.Current(); got != syntax.Equals {
		t.Fatalf("expected Equals after bump, got %s", got)
	}

	if !ts.HasPrecedingLineBreak() {
		t.Fatal("expected Equals to have a preceding line break")
	}
}

func TestTokenSourceLookaheadPastEndReturnsEOF(t *testing.T) {
	ts := NewTokenSource("foo")

	if got := ts.Nth(1); got != syntax.EOF {
		t.Fatalf("expected Nth(1) to be EOF, got %s", got)
	}

	if got := ts.Nth(10); got != syntax.EOF {
		t.Fatalf("expected lookahead past EOF to return EOF, got %s", got)
	}
}

func TestTokenSourceNegativeLookaheadPanics(t *testing.T) {
	ts := NewTokenSource("foo")

	defer func() {
		if recover() == nil {
			t.Fatal("expected negative lookahead to panic")
		}
	}()

	ts.Nth(-1)
}

func TestTokenSourceLookaheadLineBreaks(t *testing.T) {
	ts := NewTokenSource("foo bar\n# comment\nbaz qux")

	if ts.NthHasPrecedingLineBreak(0) {
		t.Fatal("foo should not have a preceding line break")
	}

	if ts.NthHasPrecedingLineBreak(1) {
		t.Fatal("bar should not have a preceding line break")
	}

	if !ts.NthHasPrecedingLineBreak(2) {
		t.Fatal("baz should have a preceding line break")
	}

	if ts.NthHasPrecedingLineBreak(3) {
		t.Fatal("qux should not have a preceding line break")
	}

	if got := ts.Current(); got != syntax.Identifier {
		t.Fatalf("lookahead advanced token source to %s", got)
	}
}
