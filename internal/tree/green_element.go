package tree

import (
	"strings"

	"github.com/arxbombus/jominia/internal/text"
)

// GreenElement is an immutable element in a green tree.
type GreenElement interface {
	Kind() RawSyntaxKind
	TextLen() text.TextSize

	writeText(*strings.Builder)
	// internal green element marker
	isGreenElement()
}

// validateGreenElement checks that the given green element is non-nil and of a valid type. Panics if the element is invalid.
func validateGreenElement(element GreenElement) {
	switch element := element.(type) {
	case *GreenNode:
		if element == nil {
			panic("tree: green node child must not be nil")
		}
	case *GreenToken:
		if element == nil {
			panic("tree: green token child must not be nil")
		}
	default:
		panic("tree: invalid green element")
	}
}
