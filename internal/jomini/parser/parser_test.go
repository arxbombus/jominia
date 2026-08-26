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
