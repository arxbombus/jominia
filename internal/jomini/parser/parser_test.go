package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
)

func TestParserBumpProducesTokenEvents(t *testing.T) {
	p := NewParser("foo = bar")

	p.Bump()
	p.Bump()
	p.Bump()

	events := p.Events()

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != EventToken {
		t.Errorf("expected first event to be EventToken")
	}

	if events[0].Kind != syntax.Identifier {
		t.Errorf("expected first token to be Identifier, got %s", events[0].Kind)
	}

	if events[0].End != 3 {
		t.Errorf("expected first token to end at 3, got %d", events[0].End)
	}

	if events[1].Kind != syntax.Equals {
		t.Errorf("expected second token to be Equals, got %s", events[1].Kind)
	}

	if events[1].End != 5 {
		t.Errorf("expected second token to end at 5, got %d", events[1].End)
	}

	if events[2].Kind != syntax.Identifier {
		t.Errorf("expected third token to be Identifier, got %s", events[2].Kind)
	}

	if events[2].End != 9 {
		t.Errorf("expected third token to end at 9, got %d", events[2].End)
	}
}

func TestParserLookaheadDoesNotAdvance(t *testing.T) {
	p := NewParser("foo # comment\n= bar")

	if got := p.Current(); got != syntax.Identifier {
		t.Fatalf("expected current token to be Identifier, got %s", got)
	}

	if got := p.Nth(0); got != syntax.Identifier {
		t.Fatalf("expected Nth(0) to be Identifier, got %s", got)
	}

	if got := p.Nth(1); got != syntax.Equals {
		t.Fatalf("expected Nth(1) to be Equals, got %s", got)
	}

	if got := p.Nth(2); got != syntax.Identifier {
		t.Fatalf("expected Nth(2) to be Identifier, got %s", got)
	}

	if got := p.Current(); got != syntax.Identifier {
		t.Fatalf("lookahead advanced parser to %s", got)
	}
}

func TestParserAt(t *testing.T) {
	p := NewParser("foo")

	if !p.At(syntax.Identifier) {
		t.Fatal("expected parser to be at Identifier")
	}

	if p.At(syntax.Equals) {
		t.Fatal("parser should not be at Equals")
	}
}

func TestParserHasPrecedingLineBreak(t *testing.T) {
	p := NewParser("foo\n= bar")

	if p.HasPrecedingLineBreak() {
		t.Fatal("first token should not have a preceding line break")
	}

	p.Bump()

	if !p.HasPrecedingLineBreak() {
		t.Fatal("Equals should have a preceding line break")
	}
}

func TestParserEat(t *testing.T) {
	p := NewParser("foo = bar")

	if p.Eat(syntax.Equals) {
		t.Fatal("Eat should not consume a non-matching token")
	}

	if !p.At(syntax.Identifier) {
		t.Fatal("failed Eat should not advance the parser")
	}

	if !p.Eat(syntax.Identifier) {
		t.Fatal("expected Eat to consume Identifier")
	}

	if !p.At(syntax.Equals) {
		t.Fatalf("expected Equals after Eat, got %s", p.Current())
	}

	events := p.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != EventToken || events[0].Kind != syntax.Identifier {
		t.Fatalf("expected Identifier token event, got %+v", events[0])
	}
}

func TestParserBumpAtEOFDoesNothing(t *testing.T) {
	p := NewParser("")

	p.Bump()

	if len(p.Events()) != 0 {
		t.Fatalf("expected no events after bumping EOF, got %d", len(p.Events()))
	}

	if !p.At(syntax.EOF) {
		t.Fatalf("expected parser to remain at EOF, got %s", p.Current())
	}
}
