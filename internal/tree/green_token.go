package tree

import (
	"strings"

	"github.com/arxbombus/jominia/internal/text"
)

// GreenToken is an immutable token in a green tree.
type GreenToken struct {
	value    string
	leading  greenTrivia
	trailing greenTrivia
	textLen  text.TextSize
	kind     RawSyntaxKind
}

// NewGreenToken returns a green token without trivia.
func NewGreenToken(kind RawSyntaxKind, value string) *GreenToken {
	return newGreenTokenWithTrivia(
		kind,
		value,
		greenTrivia{},
		greenTrivia{},
	)
}

// NewGreenTokenWithTrivia returns a green token with leading and trailing trivia.
func NewGreenTokenWithTrivia(kind RawSyntaxKind, value string, leading, trailing []TriviaPiece) *GreenToken {
	return newGreenTokenWithTrivia(
		kind,
		value,
		newGreenTrivia(leading),
		newGreenTrivia(trailing),
	)
}

// newGreenTokenWithTrivia returns a green token using already constructed green trivia.
func newGreenTokenWithTrivia(kind RawSyntaxKind, value string, leading, trailing greenTrivia) *GreenToken {
	textLen := text.SizeOf(value)
	leadingLen := leading.textLen()
	trailingLen := trailing.textLen()
	if leadingLen > textLen {
		panic("tree: leading trivia exceeds token text length")
	}
	if trailingLen > textLen-leadingLen {
		panic("tree: trivia exceeds token text length")
	}
	return &GreenToken{
		value:    strings.Clone(value), // not too sure about this one yet
		leading:  leading,
		trailing: trailing,
		textLen:  textLen,
		kind:     kind,
	}
}

// Kind returns the token's raw syntax kind.
func (t *GreenToken) Kind() RawSyntaxKind {
	return t.kind
}

// Text returns the complete source text represented by the token, including leading and trailing trivia.
func (t *GreenToken) Text() string {
	return t.value
}

// TextTrimmed returns the token text without leading or trailing trivia.
func (t *GreenToken) TextTrimmed() string {
	start := t.leading.textLen()
	end := t.textLen - t.trailing.textLen()
	return t.value[int(start):int(end)]
}

// TextLen returns the length of the complete source text represented by the token, including trivia.
func (t *GreenToken) TextLen() text.TextSize {
	return t.textLen
}

// LeadingTriviaTextLen returns the total length of the token's leading trivia.
func (t *GreenToken) LeadingTriviaTextLen() text.TextSize {
	return t.leading.textLen()
}

// TrailingTriviaTextLen returns the total length of the token's trailing trivia.
func (t *GreenToken) TrailingTriviaTextLen() text.TextSize {
	return t.trailing.textLen()
}

// LeadingTriviaCount returns the number of leading trivia pieces.
func (t *GreenToken) LeadingTriviaCount() int {
	return t.leading.len()
}

// LeadingTriviaPiece returns the leading trivia piece at index.
func (t *GreenToken) LeadingTriviaPiece(index int) TriviaPiece {
	return t.leading.piece(index)
}

// TrailingTriviaCount returns the number of trailing trivia pieces.
func (t *GreenToken) TrailingTriviaCount() int {
	return t.trailing.len()
}

// TrailingTriviaPiece returns the trailing trivia piece at index.
func (t *GreenToken) TrailingTriviaPiece(index int) TriviaPiece {
	return t.trailing.piece(index)
}

func (t *GreenToken) writeText(builder *strings.Builder) {
	builder.WriteString(t.value)
}

// internal green element marker
func (*GreenToken) isGreenElement() {}
