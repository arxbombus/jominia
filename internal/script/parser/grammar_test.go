package parser

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/tree"
)

func parseGrammarTree(t *testing.T, source string) *tree.GreenNode {
	t.Helper()

	root := Parse(source)

	if got := root.Text(); got != source {
		t.Fatalf("root text:\n%q\nwant:\n%q", got, source)
	}

	return root
}

func grammarNodesOfKind(node *tree.GreenNode, kind syntax.SyntaxKind) []*tree.GreenNode {
	var nodes []*tree.GreenNode

	var visit func(*tree.GreenNode)
	visit = func(current *tree.GreenNode) {
		if syntax.FromRaw(current.Kind()) == kind {
			nodes = append(nodes, current)
		}

		for i := 0; i < current.ChildCount(); i++ {
			child, ok := current.Child(i).(*tree.GreenNode)
			if ok {
				visit(child)
			}
		}
	}

	visit(node)
	return nodes
}

func directGrammarNodeKinds(node *tree.GreenNode) []syntax.SyntaxKind {
	var kinds []syntax.SyntaxKind

	for i := 0; i < node.ChildCount(); i++ {
		child, ok := node.Child(i).(*tree.GreenNode)
		if ok {
			kinds = append(kinds, syntax.FromRaw(child.Kind()))
		}
	}

	return kinds
}

func directGrammarTokenKinds(node *tree.GreenNode) []syntax.SyntaxKind {
	var kinds []syntax.SyntaxKind

	for i := 0; i < node.ChildCount(); i++ {
		child, ok := node.Child(i).(*tree.GreenToken)
		if ok {
			kinds = append(kinds, syntax.FromRaw(child.Kind()))
		}
	}

	return kinds
}

func grammarTokenTexts(node *tree.GreenNode) []string {
	var texts []string

	var visit func(tree.GreenElement)
	visit = func(element tree.GreenElement) {
		switch element := element.(type) {
		case *tree.GreenNode:
			for i := 0; i < element.ChildCount(); i++ {
				visit(element.Child(i))
			}
		case *tree.GreenToken:
			if syntax.FromRaw(element.Kind()) != syntax.EOF {
				texts = append(texts, element.TextTrimmed())
			}
		}
	}

	visit(node)
	return texts
}

func grammarScalarListTexts(header *tree.GreenNode) [][]string {
	var lists [][]string

	for i := 0; i < header.ChildCount(); i++ {
		child, ok := header.Child(i).(*tree.GreenNode)
		if !ok || syntax.FromRaw(child.Kind()) != syntax.ScalarList {
			continue
		}

		lists = append(lists, grammarTokenTexts(child))
	}

	return lists
}

func rootStatementList(t *testing.T, root *tree.GreenNode) *tree.GreenNode {
	t.Helper()

	if root.ChildCount() == 0 {
		t.Fatal("root has no children")
	}

	list, ok := root.Child(0).(*tree.GreenNode)
	if !ok {
		t.Fatal("first root child is not a node")
	}

	if got := syntax.FromRaw(list.Kind()); got != syntax.StatementList {
		t.Fatalf("first root child kind = %s, want StatementList", got)
	}

	return list
}

func assertRootStatementKinds(
	t *testing.T,
	root *tree.GreenNode,
	want []syntax.SyntaxKind,
) {
	t.Helper()

	got := directGrammarNodeKinds(rootStatementList(t, root))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statement kinds:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarParsesSimpleStatements(t *testing.T) {
	tests := []struct {
		name   string
		source string
		kind   syntax.SyntaxKind
	}{
		{name: "bare value", source: "optimize_memory", kind: syntax.ValueStatement},
		{name: "binary", source: "foo = bar", kind: syntax.BinaryStatement},
		{name: "comparison", source: "age >= 16", kind: syntax.BinaryStatement},
		{name: "anonymous block", source: "{}", kind: syntax.ValueStatement},
		{name: "opaque bracket", source: "[Foo(Bar)]", kind: syntax.ValueStatement},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := parseGrammarTree(t, test.source)
			assertRootStatementKinds(t, root, []syntax.SyntaxKind{test.kind})
		})
	}
}

func TestGrammarParsesBlockHeaderForms(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		scalarLists [][]string
		operators   []syntax.SyntaxKind
	}{
		{
			name:        "shorthand block",
			source:      `foo {}`,
			scalarLists: [][]string{{"foo"}},
		},
		{
			name:        "operator block",
			source:      `foo = {}`,
			scalarLists: [][]string{{"foo"}},
			operators:   []syntax.SyntaxKind{syntax.Equals},
		},
		{
			name:        "tagged block",
			source:      `color = rgb {}`,
			scalarLists: [][]string{{"color"}, {"rgb"}},
			operators:   []syntax.SyntaxKind{syntax.Equals},
		},
		{
			name: "multi-head block",
			source: `types wargoal_types
{}`,
			scalarLists: [][]string{{"types", "wargoal_types"}},
		},
		{
			name:        "quoted header argument",
			source:      `blockoverride "window_header_name" {}`,
			scalarLists: [][]string{{"blockoverride", `"window_header_name"`}},
		},
		{
			name:        "multi-head tagged block",
			source:      `type add_wargoal_panel = default_block_window {}`,
			scalarLists: [][]string{{"type", "add_wargoal_panel"}, {"default_block_window"}},
			operators:   []syntax.SyntaxKind{syntax.Equals},
		},
		{
			name: "comment before block",
			source: `foo = # comment
{}`,
			scalarLists: [][]string{{"foo"}},
			operators:   []syntax.SyntaxKind{syntax.Equals},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := parseGrammarTree(t, test.source)
			assertRootStatementKinds(t, root, []syntax.SyntaxKind{syntax.BlockStatement})

			headers := grammarNodesOfKind(root, syntax.BlockHeader)
			if len(headers) != 1 {
				t.Fatalf("block header count = %d, want 1", len(headers))
			}

			if got := grammarScalarListTexts(headers[0]); !reflect.DeepEqual(got, test.scalarLists) {
				t.Fatalf("scalar lists:\n got: %v\nwant: %v", got, test.scalarLists)
			}

			if got := directGrammarTokenKinds(headers[0]); !reflect.DeepEqual(got, test.operators) {
				t.Fatalf("header operators:\n got: %v\nwant: %v", got, test.operators)
			}
		})
	}
}

func TestGrammarParsesVic3GuiBlockHeaders(t *testing.T) {
	const source = `types wargoal_types
{
    type add_wargoal_panel = default_block_window {
        blockoverride "window_header_name" {
            text = "ADD_WARGOAL_HEADER"
        }
    }
}`

	root := parseGrammarTree(t, source)
	headers := grammarNodesOfKind(root, syntax.BlockHeader)

	if len(headers) != 3 {
		t.Fatalf("block header count = %d, want 3", len(headers))
	}

	want := [][][]string{
		{{"types", "wargoal_types"}},
		{{"type", "add_wargoal_panel"}, {"default_block_window"}},
		{{"blockoverride", `"window_header_name"`}},
	}

	for i, header := range headers {
		if got := grammarScalarListTexts(header); !reflect.DeepEqual(got, want[i]) {
			t.Errorf("block header %d:\n got: %v\nwant: %v", i, got, want[i])
		}
	}
}

func TestGrammarDoesNotMergeBlockHeaderAcrossLineBreak(t *testing.T) {
	root := parseGrammarTree(t, "optimize_memory\nfoo {}")

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.ValueStatement,
		syntax.BlockStatement,
	})
}

func TestGrammarLimitsBlockHeaderTrailingScalar(t *testing.T) {
	root := parseGrammarTree(t, `foo = red green {}`)

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BinaryStatement,
		syntax.BlockStatement,
	})

	valueLists := grammarNodesOfKind(root, syntax.ValueList)
	if len(valueLists) != 1 {
		t.Fatalf("value list count = %d, want 1", len(valueLists))
	}

	if got, want := grammarTokenTexts(valueLists[0]), []string{"red"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("value list:\n got: %v\nwant: %v", got, want)
	}

	headers := grammarNodesOfKind(root, syntax.BlockHeader)
	if len(headers) != 1 {
		t.Fatalf("block header count = %d, want 1", len(headers))
	}

	if got, want := grammarScalarListTexts(headers[0]), [][]string{{"green"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("block header scalar lists:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarParsesDenseStatements(t *testing.T) {
	root := parseGrammarTree(t, `a={b="1"c=d}foo=bar#good`)

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BlockStatement,
		syntax.BinaryStatement,
	})
}

func TestGrammarParsesMixedBlock(t *testing.T) {
	root := parseGrammarTree(t, `levels={10 0=2 1=2}`)

	lists := grammarNodesOfKind(root, syntax.StatementList)
	if len(lists) < 2 {
		t.Fatalf("statement list count = %d, want at least 2", len(lists))
	}

	got := directGrammarNodeKinds(lists[1])
	want := []syntax.SyntaxKind{
		syntax.ValueStatement,
		syntax.BinaryStatement,
		syntax.BinaryStatement,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mixed block statements:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarParsesUnbracedValueList(t *testing.T) {
	root := parseGrammarTree(t, `pattern = list "christian_emblems_list"`)

	valueLists := grammarNodesOfKind(root, syntax.ValueList)
	if len(valueLists) != 1 {
		t.Fatalf("value list count = %d, want 1", len(valueLists))
	}

	got := grammarTokenTexts(valueLists[0])
	want := []string{"list", `"christian_emblems_list"`}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("value list:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarStopsValueListBeforeNextStatement(t *testing.T) {
	root := parseGrammarTree(t, `foo = red bar = blue`)

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BinaryStatement,
		syntax.BinaryStatement,
	})
}

func TestGrammarPreservesIncompleteBinaryStatement(t *testing.T) {
	root := parseGrammarTree(t, `foo =`)

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{syntax.BinaryStatement})

	valueLists := grammarNodesOfKind(root, syntax.ValueList)
	if len(valueLists) != 1 {
		t.Fatalf("value list count = %d, want 1", len(valueLists))
	}
	if valueLists[0].ChildCount() != 0 {
		t.Fatalf("incomplete value list child count = %d, want 0", valueLists[0].ChildCount())
	}
}

func TestGrammarRecoversRepeatedOperatorAsSingleBogusStatement(t *testing.T) {
	root := parseGrammarTree(t, "bad = = foo\ngood = yes")

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BogusStatement,
		syntax.BinaryStatement,
	})

	bogus := grammarNodesOfKind(root, syntax.BogusStatement)
	if len(bogus) != 1 {
		t.Fatalf("bogus statement count = %d, want 1", len(bogus))
	}

	got := grammarTokenTexts(bogus[0])
	want := []string{"bad", "=", "=", "foo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bogus statement tokens:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarRecoversMalformedBlockHeaderAsSingleBogusStatement(t *testing.T) {
	const source = `type bad_widget = default_block_window = {
    nested = yes
}
good = yes`

	root := parseGrammarTree(t, source)

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BogusStatement,
		syntax.BinaryStatement,
	})

	bogus := grammarNodesOfKind(root, syntax.BogusStatement)
	if len(bogus) != 1 {
		t.Fatalf("bogus statement count = %d, want 1", len(bogus))
	}

	blocks := grammarNodesOfKind(bogus[0], syntax.Block)
	if len(blocks) != 1 {
		t.Fatalf("block count inside bogus statement = %d, want 1", len(blocks))
	}
}

func TestGrammarGenericRecoveryUsesLineBreakBoundary(t *testing.T) {
	root := parseGrammarTree(t, "= broken\ngood = yes")

	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BogusStatement,
		syntax.BinaryStatement,
	})

	bogus := grammarNodesOfKind(root, syntax.BogusStatement)
	if len(bogus) != 1 {
		t.Fatalf("bogus statement count = %d, want 1", len(bogus))
	}

	if got, want := grammarTokenTexts(bogus[0]), []string{"=", "broken"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bogus statement tokens:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarOpaqueGroupsRemainStructured(t *testing.T) {
	root := parseGrammarTree(t, `[MakeLineIf(IsZero(State.GetTradeCapacity), 'KEY')]`)

	if len(grammarNodesOfKind(root, syntax.BracketGroup)) != 1 {
		t.Fatal("expected one bracket group")
	}
	if len(grammarNodesOfKind(root, syntax.ParenGroup)) != 2 {
		t.Fatal("expected two parenthesis groups")
	}
}

func TestGrammarParsesConditionalBlocks(t *testing.T) {
	root := parseGrammarTree(t, `generate_advisor = {
    [[scaled_skill]
        $scaled_skill$
    ]
    [[!skill] if = {} ]
}`)

	conditionals := grammarNodesOfKind(root, syntax.ConditionalBlock)
	if len(conditionals) != 2 {
		t.Fatalf("conditional block count = %d, want 2", len(conditionals))
	}

	headers := grammarNodesOfKind(root, syntax.ConditionalHeader)
	if len(headers) != 2 {
		t.Fatalf("conditional header count = %d, want 2", len(headers))
	}
	wantHeaders := [][]string{
		{"[", "scaled_skill", "]"},
		{"[", "!", "skill", "]"},
	}
	for index, header := range headers {
		if got := grammarTokenTexts(header); !reflect.DeepEqual(got, wantHeaders[index]) {
			t.Errorf("header %d tokens:\n got: %v\nwant: %v", index, got, wantHeaders[index])
		}
	}

	firstBody := grammarNodesOfKind(conditionals[0], syntax.StatementList)
	if len(firstBody) != 1 {
		t.Fatalf("first conditional body count = %d, want 1", len(firstBody))
	}
	if got, want := directGrammarNodeKinds(firstBody[0]), []syntax.SyntaxKind{syntax.ValueStatement}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first conditional body:\n got: %v\nwant: %v", got, want)
	}

	secondBody := grammarNodesOfKind(conditionals[1], syntax.StatementList)
	if len(secondBody) != 2 {
		// The second entry is the StatementList inside if = {}.
		t.Fatalf("second conditional statement-list count = %d, want 2", len(secondBody))
	}
	if got, want := directGrammarNodeKinds(secondBody[0]), []syntax.SyntaxKind{syntax.BlockStatement}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second conditional body:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarConditionalBlockRequiresAdjacentOpeners(t *testing.T) {
	conditional := parseGrammarTree(t, `[[enabled] value]`)
	if got := len(grammarNodesOfKind(conditional, syntax.ConditionalBlock)); got != 1 {
		t.Fatalf("adjacent conditional count = %d, want 1", got)
	}

	ordinary := parseGrammarTree(t, `[ [enabled] value ]`)
	if got := len(grammarNodesOfKind(ordinary, syntax.ConditionalBlock)); got != 0 {
		t.Fatalf("spaced conditional count = %d, want 0", got)
	}
	if got := len(grammarNodesOfKind(ordinary, syntax.BracketGroup)); got != 2 {
		t.Fatalf("ordinary bracket group count = %d, want 2", got)
	}
}

func TestGrammarParsesNestedConditionalBlocks(t *testing.T) {
	root := parseGrammarTree(t, `[[outer] [[inner] value] ]`)
	if got := len(grammarNodesOfKind(root, syntax.ConditionalBlock)); got != 2 {
		t.Fatalf("conditional block count = %d, want 2", got)
	}
}

func TestGrammarParsesInlineMathPrecedence(t *testing.T) {
	root := parseGrammarTree(t, `@third = @[1/3+4*2]`)
	assertRootStatementKinds(t, root, []syntax.SyntaxKind{syntax.BinaryStatement})

	inlineMath := grammarNodesOfKind(root, syntax.InlineMath)
	if len(inlineMath) != 1 {
		t.Fatalf("inline math count = %d, want 1", len(inlineMath))
	}
	if got, want := directGrammarNodeKinds(inlineMath[0]), []syntax.SyntaxKind{syntax.BinaryExpression}; !reflect.DeepEqual(got, want) {
		t.Fatalf("inline math children:\n got: %v\nwant: %v", got, want)
	}

	expressions := grammarNodesOfKind(root, syntax.BinaryExpression)
	if len(expressions) != 3 {
		t.Fatalf("binary expression count = %d, want 3", len(expressions))
	}
	if got, want := directGrammarTokenKinds(expressions[0]), []syntax.SyntaxKind{syntax.Plus}; !reflect.DeepEqual(got, want) {
		t.Fatalf("outer operator:\n got: %v\nwant: %v", got, want)
	}
	if got, want := directGrammarTokenKinds(expressions[1]), []syntax.SyntaxKind{syntax.Slash}; !reflect.DeepEqual(got, want) {
		t.Fatalf("left operator:\n got: %v\nwant: %v", got, want)
	}
	if got, want := directGrammarTokenKinds(expressions[2]), []syntax.SyntaxKind{syntax.Star}; !reflect.DeepEqual(got, want) {
		t.Fatalf("right operator:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarInlineMathIsLeftAssociative(t *testing.T) {
	root := parseGrammarTree(t, `value = @[8/4/2]`)
	expressions := grammarNodesOfKind(root, syntax.BinaryExpression)
	if len(expressions) != 2 {
		t.Fatalf("binary expression count = %d, want 2", len(expressions))
	}
	if got, want := directGrammarNodeKinds(expressions[0]), []syntax.SyntaxKind{syntax.BinaryExpression, syntax.NumberExpression}; !reflect.DeepEqual(got, want) {
		t.Fatalf("outer expression children:\n got: %v\nwant: %v", got, want)
	}
}

func TestGrammarParsesInlineMathFactors(t *testing.T) {
	root := parseGrammarTree(t, `value = @[-(@base + |$DELTA|2$|)]`)
	wantCounts := map[syntax.SyntaxKind]int{
		syntax.InlineMath:              1,
		syntax.UnaryExpression:         1,
		syntax.ParenthesizedExpression: 1,
		syntax.NameExpression:          1,
		syntax.AbsoluteExpression:      1,
		syntax.ParameterExpression:     1,
	}
	for kind, want := range wantCounts {
		if got := len(grammarNodesOfKind(root, kind)); got != want {
			t.Errorf("%s count = %d, want %d", kind, got, want)
		}
	}
}

func TestGrammarInlineMathTreatsBooleanWordsAsNames(t *testing.T) {
	root := parseGrammarTree(t, `value = @[yes+no]`)
	if got := len(grammarNodesOfKind(root, syntax.NameExpression)); got != 2 {
		t.Fatalf("name expression count = %d, want 2", got)
	}
	if got := len(grammarNodesOfKind(root, syntax.BogusExpression)); got != 0 {
		t.Fatalf("bogus expression count = %d, want 0", got)
	}
}

func TestGrammarParsesEscapedInlineMathStart(t *testing.T) {
	root := parseGrammarTree(t, `value = @\[1+2]`)
	if got := len(grammarNodesOfKind(root, syntax.InlineMath)); got != 1 {
		t.Fatalf("inline math count = %d, want 1", got)
	}
}

func TestGrammarRecoversMalformedInlineMath(t *testing.T) {
	root := parseGrammarTree(t, "value = @[1 + ]\ngood = yes")
	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BinaryStatement,
		syntax.BinaryStatement,
	})
	if got := len(grammarNodesOfKind(root, syntax.BogusExpression)); got != 1 {
		t.Fatalf("bogus expression count = %d, want 1", got)
	}
}

func TestGrammarRecoversMalformedInlineMathParameter(t *testing.T) {
	root := parseGrammarTree(t, `value = @[$MISSING+1]`)
	if got := len(grammarNodesOfKind(root, syntax.BogusExpression)); got != 1 {
		t.Fatalf("bogus expression count = %d, want 1", got)
	}
}

func TestGrammarMissingInlineMathCloseRecoversAtLineBreak(t *testing.T) {
	root := parseGrammarTree(t, "value = @[1+2\ngood = yes")
	assertRootStatementKinds(t, root, []syntax.SyntaxKind{
		syntax.BinaryStatement,
		syntax.BinaryStatement,
	})
}

func TestGrammarExplorationCorpusIsLossless(t *testing.T) {
	source, err := os.ReadFile("testdata/syntax-exploration.txt")
	if err != nil {
		t.Fatal(err)
	}

	parseGrammarTree(t, string(source))
}

func TestGrammarLongPositionalListIsLossless(t *testing.T) {
	const count = 4096

	source := "values = { " + strings.Repeat("x ", count) + "}"
	root := parseGrammarTree(t, source)

	values := grammarNodesOfKind(root, syntax.ValueStatement)
	if len(values) != count {
		t.Fatalf("value statement count = %d, want %d", len(values), count)
	}
}
