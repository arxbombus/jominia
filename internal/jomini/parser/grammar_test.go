package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
)

type expectedEvent struct {
	eventType EventType
	kind      syntax.SyntaxKind
}

func assertGrammarEvents(
	t *testing.T,
	source string,
	expected []expectedEvent,
) {
	t.Helper()

	parser := NewParser(source)
	parseRoot(parser)

	events := parser.Events()

	if len(events) != len(expected) {
		t.Fatalf(
			"expected %d events, got %d",
			len(expected),
			len(events),
		)
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

func TestGrammarParsesSimpleEntry(t *testing.T) {
	assertGrammarEvents(t, "foo = bar", []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarParsesNestedBlock(t *testing.T) {
	assertGrammarEvents(t, `STATE_LOMBARDY = {
    id = 76
    subsistence_building = "building_subsistence_farm"
}`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},

		{eventType: EventStart, kind: syntax.Block},
		{eventType: EventToken, kind: syntax.LCurly},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.String},
		{eventType: EventFinish},

		{eventType: EventToken, kind: syntax.RCurly},
		{eventType: EventFinish},

		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarParsesBlockWithoutEquals(t *testing.T) {
	assertGrammarEvents(t, `foo {
    bar = baz
}`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},

		{eventType: EventStart, kind: syntax.Block},
		{eventType: EventToken, kind: syntax.LCurly},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventToken, kind: syntax.RCurly},
		{eventType: EventFinish},

		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarParsesBareBlockValues(t *testing.T) {
	assertGrammarEvents(t, `color = { 118 99 151 }`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},

		{eventType: EventStart, kind: syntax.Block},
		{eventType: EventToken, kind: syntax.LCurly},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventToken, kind: syntax.RCurly},
		{eventType: EventFinish},

		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}
