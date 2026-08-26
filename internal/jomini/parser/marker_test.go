package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
)

func TestMarkerCompleteCreatesNodeEvents(t *testing.T) {
	p := NewParser("foo = bar")

	m := p.Start()

	p.Bump()
	p.Bump()
	p.Bump()

	m.Complete(p, syntax.Entry)

	events := p.Events()

	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}

	if events[0].Type != EventStart {
		t.Errorf("expected first event to be EventStart")
	}

	if events[0].Kind != syntax.Entry {
		t.Errorf("expected first event kind to be Entry, got %s", events[0].Kind)
	}

	if events[1].Type != EventToken || events[1].Kind != syntax.Identifier {
		t.Errorf("expected second event to be Identifier token")
	}

	if events[2].Type != EventToken || events[2].Kind != syntax.Equals {
		t.Errorf("expected third event to be Equals token")
	}

	if events[3].Type != EventToken || events[3].Kind != syntax.Identifier {
		t.Errorf("expected fourth event to be Identifier token")
	}

	if events[4].Type != EventFinish {
		t.Errorf("expected last event to be EventFinish")
	}
}

func TestMarkerCompleteCreatesNestedNodeEvents(t *testing.T) {
	p := NewParser(`STATE_LOMBARDY = {
    id = 76
    subsistence_building = "building_subsistence_farm"
}`)

	state := p.Start()

	p.Bump() // STATE_LOMBARDY
	p.Bump() // =

	block := p.Start()

	p.Bump() // {

	id := p.Start()

	p.Bump() // id
	p.Bump() // =
	p.Bump() // 76

	id.Complete(p, syntax.Entry)

	subsistenceBuilding := p.Start()

	p.Bump() // subsistence_building
	p.Bump() // =
	p.Bump() // "building_subsistence_farm"

	subsistenceBuilding.Complete(p, syntax.Entry)

	p.Bump() // }

	block.Complete(p, syntax.Block)
	state.Complete(p, syntax.Entry)

	events := p.Events()

	if len(events) != 18 {
		t.Fatalf("expected 18 events, got %d", len(events))
	}

	expected := []struct {
		eventType EventType
		kind      syntax.SyntaxKind
	}{
		{EventStart, syntax.Entry},
		{EventToken, syntax.Identifier},
		{EventToken, syntax.Equals},

		{EventStart, syntax.Block},
		{EventToken, syntax.LCurly},

		{EventStart, syntax.Entry},
		{EventToken, syntax.Identifier},
		{EventToken, syntax.Equals},
		{EventToken, syntax.Number},
		{EventFinish, 0},

		{EventStart, syntax.Entry},
		{EventToken, syntax.Identifier},
		{EventToken, syntax.Equals},
		{EventToken, syntax.String},
		{EventFinish, 0},

		{EventToken, syntax.RCurly},
		{EventFinish, 0},
		{EventFinish, 0},
	}

	for i, want := range expected {
		got := events[i]

		if got.Type != want.eventType {
			t.Errorf(
				"event %d: expected type %d, got %d",
				i,
				want.eventType,
				got.Type,
			)
			continue
		}

		if got.Type != EventFinish && got.Kind != want.kind {
			t.Errorf(
				"event %d: expected kind %s, got %s",
				i,
				want.kind,
				got.Kind,
			)
		}
	}
}
