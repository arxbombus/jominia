package parser

import "github.com/arxbombus/jominia/internal/jomini/syntax"

// Parser consumes tokens and records syntax events.
type Parser struct {
	source *TokenSource
	events []Event
}

// NewParser returns a Parser for source.
func NewParser(source string) *Parser {
	return &Parser{
		source: NewTokenSource(source),
	}
}

// Current returns the kind of the current token.
func (p *Parser) Current() syntax.SyntaxKind {
	return p.source.Current()
}

// Nth returns the kind of the nth non-trivia token without consuming it.
// Nth(0) is equivalent to Current.
func (p *Parser) Nth(n int) syntax.SyntaxKind {
	return p.source.Nth(n)
}

// At reports whether the current token has the given kind.
func (p *Parser) At(kind syntax.SyntaxKind) bool {
	return p.Current() == kind
}

// HasPrecedingLineBreak reports whether the current token is preceded by a newline.
func (p *Parser) HasPrecedingLineBreak() bool {
	return p.source.HasPrecedingLineBreak()
}

// Start begins a new syntax node and returns a marker for its start event.
func (p *Parser) Start() Marker {
	position := len(p.events)
	p.events = append(p.events, tombstoneEvent())
	return Marker{
		position: position,
	}
}

// Bump emits the current token and advances to the next non-trivia token.
func (p *Parser) Bump() {
	kind := p.Current()
	if kind == syntax.EOF {
		return
	}
	end := p.source.CurrentRange().End()
	p.events = append(p.events, Event{
		Type: EventToken,
		Kind: kind,
		End:  end,
	})
	p.source.Bump()
}

// Eat consumes the current token if it has the given kind.
func (p *Parser) Eat(kind syntax.SyntaxKind) bool {
	if !p.At(kind) {
		return false
	}
	p.Bump()
	return true
}

// Events returns the parser's event stream.
func (p *Parser) Events() []Event {
	return p.events
}
