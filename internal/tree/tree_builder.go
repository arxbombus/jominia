package tree

type treeBuilderParent struct {
	kind       RawSyntaxKind
	firstChild int
}

// TreeBuilder incrementally constructs a green tree.
type TreeBuilder struct {
	cache    *NodeCache
	parents  []treeBuilderParent
	children []hashedGreenElement
	finished bool
}

// NewTreeBuilder returns a TreeBuilder with its own node cache.
func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{cache: NewNodeCache()}
}

// NewTreeBuilderWithCache returns a TreeBuilder that reuses cache.
func NewTreeBuilderWithCache(cache *NodeCache) *TreeBuilder {
	if cache == nil {
		panic("tree: node cache must not be nil")
	}
	cache.incrementGeneration()
	return &TreeBuilder{cache: cache}
}

// StartNode starts a new node and makes it the current node.
func (tb *TreeBuilder) StartNode(kind RawSyntaxKind) {
	tb.ensureActive()
	// only one root node is allowed
	if len(tb.parents) == 0 && len(tb.children) != 0 {
		panic("tree: cannot start another root node")
	}
	tb.parents = append(tb.parents, treeBuilderParent{
		kind:       kind,
		firstChild: len(tb.children),
	})
}

// FinishNode finishes the current node and restores its parent as the current node.
func (tb *TreeBuilder) FinishNode() {
	tb.ensureActive()
	if len(tb.parents) == 0 {
		panic("tree: no node to finish")
	}
	parentIndex := len(tb.parents) - 1
	parent := tb.parents[parentIndex]
	tb.parents = tb.parents[:parentIndex]

	childSpan := tb.children[parent.firstChild:]
	entry := tb.cache.node(parent.kind, childSpan)

	var hash uint64
	var node *GreenNode

	switch entry.state {
	case nodeCacheEntryNoCache:
		node = tb.buildNode(parent.kind, parent.firstChild)
		hash = entry.hash
	case nodeCacheEntryVacant:
		node = tb.buildNode(parent.kind, parent.firstChild)
		hash = entry.cacheNode(node)
	case nodeCacheEntryCached:
		tb.truncateChildren(parent.firstChild)
		node = entry.cached
		hash = entry.hash
	default:
		panic("tree: unexpected node cache entry state")
	}

	tb.children = append(tb.children, hashedGreenElement{
		hash:    hash,
		element: node,
	})
}

// Finish completes tree construction and returns the root green node.
func (tb *TreeBuilder) Finish() *GreenNode {
	tb.ensureActive()
	if len(tb.parents) != 0 {
		panic("tree: cannot finish tree with open nodes")
	}
	if len(tb.children) != 1 {
		panic("tree: finished tree must contain exactly one root")
	}
	root, ok := tb.children[0].element.(*GreenNode)
	if !ok || root == nil {
		panic("tree: root element must be a green node")
	}
	tb.cache.retainCache()
	tb.parents = nil
	tb.children = nil
	tb.finished = true
	return root
}

// Token adds a token without trivia to the current node.
func (tb *TreeBuilder) Token(kind RawSyntaxKind, value string) {
	tb.ensureCurrentNode()
	hash, token := tb.cache.token(kind, value)
	tb.children = append(tb.children, hashedGreenElement{
		hash:    hash,
		element: token,
	})
}

// TokenWithTrivia adds a token with leading and trailing trivia to the current node.
func (tb *TreeBuilder) TokenWithTrivia(kind RawSyntaxKind, value string, leading, trailing []TriviaPiece) {
	tb.ensureCurrentNode()
	hash, token := tb.cache.tokenWithTrivia(kind, value, leading, trailing)
	tb.children = append(tb.children, hashedGreenElement{
		hash:    hash,
		element: token,
	})
}

func (tb *TreeBuilder) buildNode(kind RawSyntaxKind, firstChild int) *GreenNode {
	childSpan := tb.children[firstChild:]
	children := make([]GreenElement, len(childSpan))
	for i, slot := range childSpan {
		children[i] = slot.element
	}
	tb.truncateChildren(firstChild)
	return newGreenNodeOwned(kind, children)
}

func (tb *TreeBuilder) truncateChildren(firstChild int) {
	clear(tb.children[firstChild:])
	tb.children = tb.children[:firstChild]
}

func (tb *TreeBuilder) ensureActive() {
	if tb == nil {
		panic("tree: nil tree builder")
	}
	if tb.finished {
		panic("tree: tree builder is already finished")
	}
}

func (tb *TreeBuilder) ensureCurrentNode() {
	tb.ensureActive()
	if len(tb.parents) == 0 {
		panic("tree: token must be added inside a node")
	}
}
