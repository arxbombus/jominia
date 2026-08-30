package tree

import (
	"testing"

	"github.com/arxbombus/jominia/internal/text"
)

func expectGreenNodePanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func TestGreenNodeStoresKindChildrenLengthAndText(t *testing.T) {
	foo := NewGreenToken(RawSyntaxKind(1), "foo ")
	equals := NewGreenToken(RawSyntaxKind(2), "= ")
	bar := NewGreenToken(RawSyntaxKind(1), "bar")
	nodeKind := RawSyntaxKind(100)

	node := NewGreenNode(nodeKind, []GreenElement{foo, equals, bar})

	if node.Kind() != nodeKind {
		t.Fatalf("Kind() = %v, want %v", node.Kind(), nodeKind)
	}
	if node.ChildCount() != 3 {
		t.Fatalf("ChildCount() = %d, want 3", node.ChildCount())
	}
	if node.Child(0) != foo || node.Child(1) != equals || node.Child(2) != bar {
		t.Fatal("children were not preserved in order")
	}

	const wantText = "foo = bar"
	if node.Text() != wantText {
		t.Fatalf("Text() = %q, want %q", node.Text(), wantText)
	}
	if node.TextLen() != text.SizeOf(wantText) {
		t.Fatalf("TextLen() = %d, want %d", node.TextLen(), text.SizeOf(wantText))
	}
}

func TestGreenNodeCopiesChildSlice(t *testing.T) {
	first := NewGreenToken(RawSyntaxKind(1), "first")
	second := NewGreenToken(RawSyntaxKind(1), "second")
	children := []GreenElement{first}

	node := NewGreenNode(RawSyntaxKind(100), children)
	children[0] = second

	if node.Child(0) != first {
		t.Fatal("node changed after caller mutated the original child slice")
	}
	if node.Text() != "first" {
		t.Fatalf("Text() = %q, want %q", node.Text(), "first")
	}
}

func TestGreenNodeChildOutOfRangePanics(t *testing.T) {
	node := NewGreenNode(RawSyntaxKind(100), nil)

	expectGreenNodePanic(t, func() {
		_ = node.Child(0)
	})
}

func TestGreenNodeRejectsTextLengthOverflow(t *testing.T) {
	maxLengthChild := &GreenNode{
		textLen: ^text.TextSize(0),
		kind:    RawSyntaxKind(101),
	}
	oneByteChild := NewGreenToken(RawSyntaxKind(1), "x")

	expectGreenNodePanic(t, func() {
		_ = NewGreenNode(
			RawSyntaxKind(100),
			[]GreenElement{maxLengthChild, oneByteChild},
		)
	})
}
