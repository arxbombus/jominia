package tree

import (
	"strings"
	"testing"

	"github.com/arxbombus/jominia/internal/text"
)

var (
	_ GreenElement = (*GreenNode)(nil)
	_ GreenElement = (*GreenToken)(nil)
)

type testInvalidGreenElement struct{}

func (testInvalidGreenElement) Kind() RawSyntaxKind        { return 0 }
func (testInvalidGreenElement) TextLen() text.TextSize     { return 0 }
func (testInvalidGreenElement) writeText(*strings.Builder) {}
func (testInvalidGreenElement) isGreenElement()            {}

func expectGreenElementPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	fn()
}

func TestValidateGreenElementAcceptsNodeAndToken(t *testing.T) {
	token := NewGreenToken(RawSyntaxKind(1), "foo")
	node := NewGreenNode(RawSyntaxKind(2), []GreenElement{token})

	validateGreenElement(token)
	validateGreenElement(node)
}

func TestValidateGreenElementRejectsTypedNil(t *testing.T) {
	t.Run("token", func(t *testing.T) {
		var token *GreenToken
		expectGreenElementPanic(t, func() {
			validateGreenElement(token)
		})
	})

	t.Run("node", func(t *testing.T) {
		var node *GreenNode
		expectGreenElementPanic(t, func() {
			validateGreenElement(node)
		})
	})
}

func TestValidateGreenElementRejectsUnexpectedImplementation(t *testing.T) {
	expectGreenElementPanic(t, func() {
		validateGreenElement(testInvalidGreenElement{})
	})
}
