package tree

import "testing"

func expectTreeBuilderPanic(t *testing.T, fn func()) {
	t.Helper()

	didPanic := false

	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()

		fn()
	}()

	if !didPanic {
		t.Fatal("expected panic")
	}
}

func TestTreeBuilderBuildsNestedTree(t *testing.T) {
	const (
		identifier = RawSyntaxKind(1)
		equals     = RawSyntaxKind(2)
		rootKind   = RawSyntaxKind(100)
		entryKind  = RawSyntaxKind(101)
	)

	builder := NewTreeBuilder()
	builder.StartNode(rootKind)
	builder.StartNode(entryKind)
	builder.Token(identifier, "foo")
	builder.Token(equals, " = ")
	builder.Token(identifier, "bar")
	builder.FinishNode()
	builder.FinishNode()

	root := builder.Finish()
	if root.Kind() != rootKind {
		t.Fatalf("root kind = %v, want %v", root.Kind(), rootKind)
	}
	if root.Text() != "foo = bar" {
		t.Fatalf("root text = %q, want %q", root.Text(), "foo = bar")
	}
	if root.ChildCount() != 1 {
		t.Fatalf("root child count = %d, want 1", root.ChildCount())
	}

	entry, ok := root.Child(0).(*GreenNode)
	if !ok {
		t.Fatal("root child is not a green node")
	}
	if entry.Kind() != entryKind {
		t.Fatalf("entry kind = %v, want %v", entry.Kind(), entryKind)
	}
	if entry.ChildCount() != 3 {
		t.Fatalf("entry child count = %d, want 3", entry.ChildCount())
	}
}

func TestTreeBuilderBuildsTokenWithTrivia(t *testing.T) {
	const (
		identifier = RawSyntaxKind(1)
		rootKind   = RawSyntaxKind(100)
	)

	leading := []TriviaPiece{NewWhitespaceTriviaPiece(2)}
	trailing := []TriviaPiece{
		NewWhitespaceTriviaPiece(1),
		NewCommentTriviaPiece(4),
	}

	builder := NewTreeBuilder()
	builder.StartNode(rootKind)
	builder.TokenWithTrivia(identifier, "  foo # hi", leading, trailing)
	builder.FinishNode()

	root := builder.Finish()
	token, ok := root.Child(0).(*GreenToken)
	if !ok {
		t.Fatal("root child is not a green token")
	}
	if token.Text() != "  foo # hi" {
		t.Fatalf("token text = %q, want %q", token.Text(), "  foo # hi")
	}
	if token.TextTrimmed() != "foo" {
		t.Fatalf("trimmed token text = %q, want %q", token.TextTrimmed(), "foo")
	}
	if token.LeadingTriviaCount() != 1 {
		t.Fatalf("leading trivia count = %d, want 1", token.LeadingTriviaCount())
	}
	if token.TrailingTriviaCount() != 2 {
		t.Fatalf("trailing trivia count = %d, want 2", token.TrailingTriviaCount())
	}
}

func TestTreeBuilderReusesCachedNodesInsideTree(t *testing.T) {
	const (
		identifier = RawSyntaxKind(1)
		rootKind   = RawSyntaxKind(100)
		entryKind  = RawSyntaxKind(101)
	)

	builder := NewTreeBuilder()
	builder.StartNode(rootKind)

	builder.StartNode(entryKind)
	builder.Token(identifier, "foo")
	builder.FinishNode()

	builder.StartNode(entryKind)
	builder.Token(identifier, "foo")
	builder.FinishNode()

	builder.FinishNode()
	root := builder.Finish()

	first, ok := root.Child(0).(*GreenNode)
	if !ok {
		t.Fatal("first child is not a green node")
	}
	second, ok := root.Child(1).(*GreenNode)
	if !ok {
		t.Fatal("second child is not a green node")
	}
	if first != second {
		t.Fatal("identical nodes were not reused from the cache")
	}
}

func TestTreeBuilderReusesCacheAcrossBuilders(t *testing.T) {
	const (
		identifier = RawSyntaxKind(1)
		rootKind   = RawSyntaxKind(100)
	)

	cache := NewNodeCache()

	firstBuilder := NewTreeBuilderWithCache(cache)
	firstBuilder.StartNode(rootKind)
	firstBuilder.Token(identifier, "foo")
	firstBuilder.FinishNode()
	firstRoot := firstBuilder.Finish()

	secondBuilder := NewTreeBuilderWithCache(cache)
	secondBuilder.StartNode(rootKind)
	secondBuilder.Token(identifier, "foo")
	secondBuilder.FinishNode()
	secondRoot := secondBuilder.Finish()

	if firstRoot != secondRoot {
		t.Fatal("identical roots built with the same cache were not reused")
	}
}

func TestTreeBuilderBuildsUncachedLargeNode(t *testing.T) {
	const (
		identifier = RawSyntaxKind(1)
		rootKind   = RawSyntaxKind(100)
	)

	builder := NewTreeBuilder()
	builder.StartNode(rootKind)
	builder.Token(identifier, "a")
	builder.Token(identifier, "b")
	builder.Token(identifier, "c")
	builder.Token(identifier, "d")
	builder.FinishNode()

	root := builder.Finish()
	if root.Text() != "abcd" {
		t.Fatalf("root text = %q, want %q", root.Text(), "abcd")
	}
	if root.ChildCount() != 4 {
		t.Fatalf("root child count = %d, want 4", root.ChildCount())
	}
}

func TestTreeBuilderBuildsEmptyNode(t *testing.T) {
	const rootKind = RawSyntaxKind(100)

	builder := NewTreeBuilder()
	builder.StartNode(rootKind)
	builder.FinishNode()

	root := builder.Finish()
	if root.Kind() != rootKind {
		t.Fatalf("root kind = %v, want %v", root.Kind(), rootKind)
	}
	if root.ChildCount() != 0 {
		t.Fatalf("root child count = %d, want 0", root.ChildCount())
	}
	if root.Text() != "" {
		t.Fatalf("root text = %q, want empty text", root.Text())
	}
}

func TestNewTreeBuilderWithCacheRejectsNilCache(t *testing.T) {
	expectTreeBuilderPanic(t, func() {
		_ = NewTreeBuilderWithCache(nil)
	})
}

func TestTreeBuilderRejectsTokenWithoutNode(t *testing.T) {
	builder := NewTreeBuilder()

	expectTreeBuilderPanic(t, func() {
		builder.Token(RawSyntaxKind(1), "foo")
	})
}

func TestTreeBuilderRejectsFinishNodeWithoutOpenNode(t *testing.T) {
	builder := NewTreeBuilder()

	expectTreeBuilderPanic(t, func() {
		builder.FinishNode()
	})
}

func TestTreeBuilderRejectsFinishWithOpenNode(t *testing.T) {
	builder := NewTreeBuilder()
	builder.StartNode(RawSyntaxKind(100))

	expectTreeBuilderPanic(t, func() {
		builder.Finish()
	})
}

func TestTreeBuilderRejectsFinishWithoutRoot(t *testing.T) {
	builder := NewTreeBuilder()

	expectTreeBuilderPanic(t, func() {
		builder.Finish()
	})
}

func TestTreeBuilderRejectsSecondRoot(t *testing.T) {
	builder := NewTreeBuilder()
	builder.StartNode(RawSyntaxKind(100))
	builder.FinishNode()

	expectTreeBuilderPanic(t, func() {
		builder.StartNode(RawSyntaxKind(101))
	})
}

func TestTreeBuilderRejectsUseAfterFinish(t *testing.T) {
	builder := NewTreeBuilder()
	builder.StartNode(RawSyntaxKind(100))
	builder.FinishNode()
	_ = builder.Finish()

	expectTreeBuilderPanic(t, func() {
		builder.StartNode(RawSyntaxKind(101))
	})
}
