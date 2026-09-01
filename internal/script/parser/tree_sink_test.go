package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/tree"
)

func parseGreenTree(source string) *tree.GreenNode {
	parser := NewParser(source)
	parseRoot(parser)

	events, trivia := parser.Finish()

	sink := NewLosslessTreeSink(source, trivia)
	processEvents(sink, events)

	return sink.Finish()
}

func collectGreenTokens(node *tree.GreenNode) []*tree.GreenToken {
	var tokens []*tree.GreenToken

	var visit func(tree.GreenElement)
	visit = func(element tree.GreenElement) {
		switch element := element.(type) {
		case *tree.GreenNode:
			for i := 0; i < element.ChildCount(); i++ {
				visit(element.Child(i))
			}

		case *tree.GreenToken:
			tokens = append(tokens, element)
		}
	}

	visit(node)

	return tokens
}

func findGreenToken(
	t *testing.T,
	root *tree.GreenNode,
	trimmedText string,
) *tree.GreenToken {
	t.Helper()

	for _, token := range collectGreenTokens(root) {
		if token.TextTrimmed() == trimmedText {
			return token
		}
	}

	t.Fatalf("token %q not found", trimmedText)
	return nil
}

func TestLosslessTreeSinkPreservesSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty",
			source: "",
		},
		{
			name:   "whitespace only",
			source: "   ",
		},
		{
			name:   "newlines only",
			source: "\n\n",
		},
		{
			name:   "comment only",
			source: "# comment",
		},
		{
			name:   "simple entry",
			source: "foo = bar",
		},
		{
			name:   "leading whitespace",
			source: "    foo = bar",
		},
		{
			name:   "trailing whitespace",
			source: "foo = bar    ",
		},
		{
			name:   "trailing newline",
			source: "foo = bar\n",
		},
		{
			name:   "trailing comment",
			source: "foo = bar # comment",
		},
		{
			name:   "crlf",
			source: "foo = bar\r\nbaz = qux\r\n",
		},
		{
			name:   "boundary free",
			source: `a={b="1"c=d}foo=bar#good`,
		},
		{
			name: "comment before block",
			source: `my_obj = # this is going to be great
{ # my_key = prev_value
    my_key = value # better_value
    a = "not # a comment"
} # the end`,
		},
		{
			name: "nested blocks",
			source: `foo = {
    bar = {
        baz = qux
    }
}`,
		},
		{
			name: "mixed block",
			source: `brittany_area = {
    color = { 118 99 151 }
    169 170 171 172 4384
}`,
		},
		{
			name: "representative state",
			source: `STATE_CALIFORNIA = {
    state_id = 10781
    state_trait = { state_trait_1 state_trait_2 }
    background = "gfx/icons/backgrounds/california.png"
    provinces = { x123 x234 x456 }
    arable_land = 1000
}`,
		},
		{
			name: "unclosed block",
			source: `STATE_CALIFORNIA = {
    state_id = 10781
    provinces = { x123 x234 x456 }`,
		},
		{
			name:   "tagged block",
			source: `color = rgb { 100 200 150 }`,
		},
		{
			name: "unmarked list",
			source: `simple_cross_flag = {
    pattern = list "christian_emblems_list"
    color1 = list "normal_colors"
}`,
		},
		{
			name: "vic3 gui multi head",
			source: `types wargoal_types
{
    type add_wargoal_panel = default_block_window {
        name = "add_wargoal_panel"
    }
}`,
		},
		{
			name: "eu4 parameter syntax",
			source: `generate_advisor = {
    [[scaled_skill]
        $scaled_skill$
    ]
    [[!skill] if = {} ]
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := parseGreenTree(test.source)

			if got := root.Text(); got != test.source {
				t.Fatalf(
					"root text:\n%q\nwant:\n%q",
					got,
					test.source,
				)
			}
		})
	}
}

func TestLosslessTreeSinkAttachesTokenTrivia(t *testing.T) {
	const source = "  foo = bar # comment\n    baz = qux"

	root := parseGreenTree(source)

	foo := findGreenToken(t, root, "foo")

	if foo.Text() != "  foo " {
		t.Fatalf("foo text = %q, want %q", foo.Text(), "  foo ")
	}

	if foo.LeadingTriviaCount() != 1 {
		t.Fatalf(
			"foo leading trivia count = %d, want 1",
			foo.LeadingTriviaCount(),
		)
	}

	fooLeading := foo.LeadingTriviaPiece(0)
	if !fooLeading.Kind().IsWhitespace() {
		t.Fatalf("foo leading trivia is not whitespace")
	}

	if fooLeading.TextLen() != 2 {
		t.Fatalf(
			"foo leading trivia length = %d, want 2",
			fooLeading.TextLen(),
		)
	}

	if foo.TrailingTriviaCount() != 1 {
		t.Fatalf(
			"foo trailing trivia count = %d, want 1",
			foo.TrailingTriviaCount(),
		)
	}

	fooTrailing := foo.TrailingTriviaPiece(0)
	if !fooTrailing.Kind().IsWhitespace() {
		t.Fatalf("foo trailing trivia is not whitespace")
	}

	if fooTrailing.TextLen() != 1 {
		t.Fatalf(
			"foo trailing trivia length = %d, want 1",
			fooTrailing.TextLen(),
		)
	}

	bar := findGreenToken(t, root, "bar")

	if bar.Text() != "bar # comment" {
		t.Fatalf(
			"bar text = %q, want %q",
			bar.Text(),
			"bar # comment",
		)
	}

	if bar.LeadingTriviaCount() != 0 {
		t.Fatalf(
			"bar leading trivia count = %d, want 0",
			bar.LeadingTriviaCount(),
		)
	}

	if bar.TrailingTriviaCount() != 2 {
		t.Fatalf(
			"bar trailing trivia count = %d, want 2",
			bar.TrailingTriviaCount(),
		)
	}

	if !bar.TrailingTriviaPiece(0).Kind().IsWhitespace() {
		t.Fatal("first bar trailing trivia is not whitespace")
	}

	if !bar.TrailingTriviaPiece(1).Kind().IsComment() {
		t.Fatal("second bar trailing trivia is not a comment")
	}

	baz := findGreenToken(t, root, "baz")

	if baz.Text() != "\n    baz " {
		t.Fatalf(
			"baz text = %q, want %q",
			baz.Text(),
			"\n    baz ",
		)
	}

	if baz.LeadingTriviaCount() != 2 {
		t.Fatalf(
			"baz leading trivia count = %d, want 2",
			baz.LeadingTriviaCount(),
		)
	}

	if !baz.LeadingTriviaPiece(0).Kind().IsNewline() {
		t.Fatal("first baz leading trivia is not a newline")
	}

	if !baz.LeadingTriviaPiece(1).Kind().IsWhitespace() {
		t.Fatal("second baz leading trivia is not whitespace")
	}

	if baz.TrailingTriviaCount() != 1 {
		t.Fatalf(
			"baz trailing trivia count = %d, want 1",
			baz.TrailingTriviaCount(),
		)
	}
}

func TestLosslessTreeSinkSynthesizesEOF(t *testing.T) {
	const source = "foo = bar"

	root := parseGreenTree(source)

	if root.ChildCount() == 0 {
		t.Fatal("root has no children")
	}

	last := root.Child(root.ChildCount() - 1)

	eof, ok := last.(*tree.GreenToken)
	if !ok {
		t.Fatal("last root child is not a token")
	}

	if got := syntax.FromRaw(eof.Kind()); got != syntax.EOF {
		t.Fatalf("last token kind = %s, want EOF", got)
	}

	if eof.Text() != "" {
		t.Fatalf("EOF text = %q, want empty text", eof.Text())
	}

	if eof.TextTrimmed() != "" {
		t.Fatalf(
			"EOF trimmed text = %q, want empty text",
			eof.TextTrimmed(),
		)
	}
}

func TestLosslessTreeSinkAttachesFinalTriviaToEOF(t *testing.T) {
	const source = "foo = bar\n    # trailing comment\n"

	root := parseGreenTree(source)

	last := root.Child(root.ChildCount() - 1)

	eof, ok := last.(*tree.GreenToken)
	if !ok {
		t.Fatal("last root child is not a token")
	}

	if got := syntax.FromRaw(eof.Kind()); got != syntax.EOF {
		t.Fatalf("last token kind = %s, want EOF", got)
	}

	const expectedText = "\n    # trailing comment\n"

	if eof.Text() != expectedText {
		t.Fatalf(
			"EOF text = %q, want %q",
			eof.Text(),
			expectedText,
		)
	}

	if eof.TextTrimmed() != "" {
		t.Fatalf(
			"EOF trimmed text = %q, want empty text",
			eof.TextTrimmed(),
		)
	}

	if eof.LeadingTriviaCount() != 4 {
		t.Fatalf(
			"EOF leading trivia count = %d, want 4",
			eof.LeadingTriviaCount(),
		)
	}

	if !eof.LeadingTriviaPiece(0).Kind().IsNewline() {
		t.Fatal("first EOF trivia is not a newline")
	}

	if !eof.LeadingTriviaPiece(1).Kind().IsWhitespace() {
		t.Fatal("second EOF trivia is not whitespace")
	}

	if !eof.LeadingTriviaPiece(2).Kind().IsComment() {
		t.Fatal("third EOF trivia is not a comment")
	}

	if !eof.LeadingTriviaPiece(3).Kind().IsNewline() {
		t.Fatal("fourth EOF trivia is not a newline")
	}

	if eof.TrailingTriviaCount() != 0 {
		t.Fatalf(
			"EOF trailing trivia count = %d, want 0",
			eof.TrailingTriviaCount(),
		)
	}
}

func TestLosslessTreeSinkPreservesTriviaOnlySource(t *testing.T) {
	const source = "  \n# comment\n"

	root := parseGreenTree(source)

	if root.ChildCount() != 1 {
		t.Fatalf(
			"root child count = %d, want 1",
			root.ChildCount(),
		)
	}

	eof, ok := root.Child(0).(*tree.GreenToken)
	if !ok {
		t.Fatal("root child is not a token")
	}

	if got := syntax.FromRaw(eof.Kind()); got != syntax.EOF {
		t.Fatalf("token kind = %s, want EOF", got)
	}

	if eof.Text() != source {
		t.Fatalf(
			"EOF text = %q, want %q",
			eof.Text(),
			source,
		)
	}

	if eof.TextTrimmed() != "" {
		t.Fatalf(
			"EOF trimmed text = %q, want empty text",
			eof.TextTrimmed(),
		)
	}

	if root.Text() != source {
		t.Fatalf(
			"root text = %q, want %q",
			root.Text(),
			source,
		)
	}
}

func TestLosslessTreeSinkBuildsRepresentativeTree(t *testing.T) {
	const source = `STATE_CALIFORNIA = {
    state_id = 10781
    state_trait = { state_trait_1 state_trait_2 }
    background = "gfx/icons/backgrounds/california.png"
    provinces = { x123 x234 x456 }
    arable_land = 1000
}`

	root := parseGreenTree(source)

	if got := syntax.FromRaw(root.Kind()); got != syntax.Root {
		t.Fatalf("root kind = %s, want Root", got)
	}

	if root.Text() != source {
		t.Fatalf(
			"root text does not match source:\n%s",
			root.Text(),
		)
	}

	if root.ChildCount() != 2 {
		t.Fatalf(
			"root child count = %d, want 2",
			root.ChildCount(),
		)
	}

	entry, ok := root.Child(0).(*tree.GreenNode)
	if !ok {
		t.Fatal("first root child is not a node")
	}

	if got := syntax.FromRaw(entry.Kind()); got != syntax.Entry {
		t.Fatalf(
			"first root child kind = %s, want Entry",
			got,
		)
	}

	eof, ok := root.Child(1).(*tree.GreenToken)
	if !ok {
		t.Fatal("second root child is not a token")
	}

	if got := syntax.FromRaw(eof.Kind()); got != syntax.EOF {
		t.Fatalf(
			"second root child kind = %s, want EOF",
			got,
		)
	}
}
