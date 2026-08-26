package parser

import "github.com/arxbombus/jominia/internal/jomini/syntax"

type Parser struct {
	source *TokenSource
	events []Event
}

func NewParser(source string) *Parser {
	return &Parser{
		source: NewTokenSource(source),
	}
}

func (p *Parser) Current() syntax.SyntaxKind {
	return p.source.Current()
}

func (p *Parser) At(kind syntax.SyntaxKind) bool {
	return p.Current() == kind
}

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

func (p *Parser) Eat(kind syntax.SyntaxKind) bool {
	if !p.At(kind) {
		return false
	}
	p.Bump()
	return true
}

func (p *Parser) Events() []Event {
	return p.events
}
