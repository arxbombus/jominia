package tree

import (
	"strings"

	"github.com/arxbombus/jominia/internal/text"
)

// GreenNode is an immutable node in a green tree.
type GreenNode struct {
	children []GreenElement
	textLen  text.TextSize
	kind     RawSyntaxKind
}

// NewGreenNode returns a green node with the given kind and children.
func NewGreenNode(kind RawSyntaxKind, children []GreenElement) *GreenNode {
	ownedChildren := make([]GreenElement, len(children))
	copy(ownedChildren, children)
	var textLen text.TextSize
	for _, child := range ownedChildren {
		validateGreenElement(child)
		childLen := child.TextLen()
		if childLen > ^text.TextSize(0)-textLen {
			panic("tree: node text exceeds maximum TextSize")
		}
		textLen += childLen
	}
	return &GreenNode{
		children: ownedChildren,
		textLen:  textLen,
		kind:     kind,
	}
}

// Kind returns the node's raw syntax kind.
func (n *GreenNode) Kind() RawSyntaxKind {
	return n.kind
}

// TextLen returns the total length of the source text covered by the node.
func (n *GreenNode) TextLen() text.TextSize {
	return n.textLen
}

// ChildCount returns the number of children in the node.
func (n *GreenNode) ChildCount() int {
	return len(n.children)
}

// Child returns the child at index.
func (n *GreenNode) Child(index int) GreenElement {
	return n.children[index]
}

// Text returns the complete source text covered by the node.
func (n *GreenNode) Text() string {
	var builder strings.Builder
	n.writeText(&builder)
	return builder.String()
}

func (n *GreenNode) writeText(builder *strings.Builder) {
	for _, child := range n.children {
		child.writeText(builder)
	}
}

// internal green element marker
func (n *GreenNode) isGreenElement() {}
