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

func TestGrammarParsesOperators(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		operator  syntax.SyntaxKind
		valueKind syntax.SyntaxKind
	}{
		{"equals", "foo = bar", syntax.Equals, syntax.Identifier},
		{"equals equals", "foo == bar", syntax.EqualsEquals, syntax.Identifier},
		{"not equals", "foo != bar", syntax.BangEquals, syntax.Identifier},
		{"less", "count < 2", syntax.Less, syntax.Number},
		{"less equals", "count <= 2", syntax.LessEquals, syntax.Number},
		{"greater", "age > 16", syntax.Greater, syntax.Number},
		{"greater equals", "age >= 16", syntax.GreaterEquals, syntax.Number},
		{"question equals", "c:RUS ?= this", syntax.QuestionEquals, syntax.Identifier},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertGrammarEvents(t, test.source, []expectedEvent{
				{eventType: EventStart, kind: syntax.Root},
				{eventType: EventStart, kind: syntax.Entry},
				{eventType: EventToken, kind: syntax.Identifier},
				{eventType: EventToken, kind: test.operator},
				{eventType: EventToken, kind: test.valueKind},
				{eventType: EventFinish},
				{eventType: EventFinish},
			})
		})
	}
}

func TestGrammarParsesValueSequence(t *testing.T) {
	assertGrammarEvents(t, `pattern = list "christian_emblems_list"`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.String},
		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarStopsValueSequenceAtNextEntry(t *testing.T) {
	assertGrammarEvents(t, `a=1 b=2 c=3`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventFinish},

		{eventType: EventFinish},
	})
}

func TestGrammarStopsValueSequenceAtLineBreak(t *testing.T) {
	assertGrammarEvents(t, "foo = red\nbar = blue", []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventFinish},
	})
}

func TestGrammarParsesTaggedBlockValue(t *testing.T) {
	assertGrammarEvents(t, `color = rgb { 100 200 150 }`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},

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

func TestGrammarDoesNotUseMultiTokenHeads(t *testing.T) {
	assertGrammarEvents(t, "types wargoal_types\n{\n}", []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventStart, kind: syntax.Block},
		{eventType: EventToken, kind: syntax.LCurly},
		{eventType: EventToken, kind: syntax.RCurly},
		{eventType: EventFinish},
		{eventType: EventFinish},

		{eventType: EventFinish},
	})
}

func TestGrammarParsesBareBlockScalarsAsSeparateEntries(t *testing.T) {
	assertGrammarEvents(t, `foo = { red green blue = yes }`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},

		{eventType: EventStart, kind: syntax.Block},
		{eventType: EventToken, kind: syntax.LCurly},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

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

func TestGrammarParsesOpaqueBracketAndParenGroups(t *testing.T) {
	assertGrammarEvents(t, `[Localize( 'NEWLINE' )]`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},

		{eventType: EventStart, kind: syntax.BracketGroup},
		{eventType: EventToken, kind: syntax.LBracket},
		{eventType: EventToken, kind: syntax.Identifier},

		{eventType: EventStart, kind: syntax.ParenGroup},
		{eventType: EventToken, kind: syntax.LParen},
		{eventType: EventToken, kind: syntax.SingleQuotedString},
		{eventType: EventToken, kind: syntax.RParen},
		{eventType: EventFinish},

		{eventType: EventToken, kind: syntax.RBracket},
		{eventType: EventFinish},

		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarParsesBracketGroupAfterScalarValue(t *testing.T) {
	assertGrammarEvents(t, `@canton_scale_cross_x = @[ ( 333 / 768 ) + 0.001 ]`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},

		{eventType: EventStart, kind: syntax.BracketGroup},
		{eventType: EventToken, kind: syntax.LBracket},

		{eventType: EventStart, kind: syntax.ParenGroup},
		{eventType: EventToken, kind: syntax.LParen},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventToken, kind: syntax.RParen},
		{eventType: EventFinish},

		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Number},
		{eventType: EventToken, kind: syntax.RBracket},
		{eventType: EventFinish},

		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarTreatsSingleQuotedStringAsBogusScalarValue(t *testing.T) {
	assertGrammarEvents(t, `foo = 'bar'`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Bogus},
		{eventType: EventToken, kind: syntax.SingleQuotedString},
		{eventType: EventFinish},

		{eventType: EventFinish},
	})
}

func TestGrammarConsumesSemicolonInsideEntry(t *testing.T) {
	assertGrammarEvents(t, `textureFile3 = "gfx/foo.dds";`, []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},
		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.String},
		{eventType: EventToken, kind: syntax.Semicolon},
		{eventType: EventFinish},
		{eventType: EventFinish},
	})
}

func TestGrammarRecoversUnexpectedToken(t *testing.T) {
	assertGrammarEvents(t, "foo = bar\n)\nbaz = qux", []expectedEvent{
		{eventType: EventStart, kind: syntax.Root},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Bogus},
		{eventType: EventToken, kind: syntax.RParen},
		{eventType: EventFinish},

		{eventType: EventStart, kind: syntax.Entry},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventToken, kind: syntax.Equals},
		{eventType: EventToken, kind: syntax.Identifier},
		{eventType: EventFinish},

		{eventType: EventFinish},
	})
}
