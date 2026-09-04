package parser

import (
	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
)

const (
	inlineMathLowestPrecedence         = 1
	inlineMathAdditivePrecedence       = 1
	inlineMathMultiplicativePrecedence = 2
)

// parseInlineMath parses an @[...] (or @\[...]) expression value. The opener selects the arithmetic lexer context for the following token; the closer returns tokenization to normal script.
func parseInlineMath(parser *Parser) {
	inlineMath := parser.Start()
	if !parser.EatWithContext(syntax.InlineMathStart, lexer.LexInlineMath) {
		panic("grammar(script): expected inline math opener")
	}
	if !isInlineMathEnd(parser) {
		if _, ok := parseInlineMathExpression(parser, inlineMathLowestPrecedence); !ok {
			parseBogusInlineMathExpression(parser)
		}
	}
	if !isInlineMathEnd(parser) {
		parseBogusInlineMathExpression(parser)
	}
	if !parser.EatWithContext(syntax.RBracket, lexer.LexNormal) && !parser.At(syntax.EOF) {
		// The inline lexer normally returns to normal mode when it emits ]. If recovery stopped at another boundary, restore normal tokenization for the unconsumed token and the statements that follow it.
		parser.ReLex(lexer.ReLexNormal)
	}
	inlineMath.Complete(parser, syntax.InlineMath)
}

// parseInlineMathExpression is a precedence-climbing expression parser. Its left-associative wrapping uses CompletedMarker.Precede.
func parseInlineMathExpression(parser *Parser, minimumPrecedence int) (CompletedMarker, bool) {
	left, ok := parseInlineMathUnaryExpression(parser)
	if !ok {
		return CompletedMarker{}, false
	}
	for !parser.HasPrecedingLineBreak() {
		precedence := inlineMathBinaryPrecedence(parser.Current())
		if precedence < minimumPrecedence {
			break
		}

		binary := left.Precede(parser)
		parser.BumpWithContext(lexer.LexInlineMath)
		_, hasRight := parseInlineMathExpression(parser, precedence+1)
		kind := syntax.BinaryExpression
		if !hasRight {
			kind = syntax.BogusExpression
		}
		left = binary.Complete(parser, kind)
	}
	return left, true
}

func parseInlineMathUnaryExpression(parser *Parser) (CompletedMarker, bool) {
	if parser.At(syntax.Plus) || parser.At(syntax.Minus) {
		unary := parser.Start()
		parser.BumpWithContext(lexer.LexInlineMath)
		_, hasOperand := parseInlineMathUnaryExpression(parser)
		kind := syntax.UnaryExpression
		if !hasOperand {
			kind = syntax.BogusExpression
		}
		return unary.Complete(parser, kind), true
	}
	return parseInlineMathPrimaryExpression(parser)
}

func parseInlineMathPrimaryExpression(parser *Parser) (CompletedMarker, bool) {
	switch {
	case parser.At(syntax.Number):
		number := parser.Start()
		parser.BumpWithContext(lexer.LexInlineMath)
		return number.Complete(parser, syntax.NumberExpression), true
	case parser.At(syntax.Identifier):
		return parseInlineMathNameExpression(parser), true
	case parser.At(syntax.At):
		return parseVariableReference(parser), true
	case parser.At(syntax.Dollar):
		return parseParameterExpression(parser), true
	case parser.At(syntax.LParen):
		return parseInlineMathParenthesizedExpression(parser), true
	case parser.At(syntax.Pipe):
		return parseInlineMathAbsoluteExpression(parser), true
	default:
		return CompletedMarker{}, false
	}
}

func parseInlineMathNameExpression(parser *Parser) CompletedMarker {
	name := parser.Start()
	hasName := parser.EatWithContext(syntax.Identifier, lexer.LexInlineMath)
	kind := syntax.NameExpression
	if !hasName {
		kind = syntax.BogusExpression
	}
	return name.Complete(parser, kind)
}

func parseInlineMathParenthesizedExpression(parser *Parser) CompletedMarker {
	parenthesized := parser.Start()
	if !parser.EatWithContext(syntax.LParen, lexer.LexInlineMath) {
		panic("grammar(script): expected opening parenthesis")
	}
	_, hasExpression := parseInlineMathExpression(parser, inlineMathLowestPrecedence)
	closed := parser.EatWithContext(syntax.RParen, lexer.LexInlineMath)
	kind := syntax.ParenthesizedExpression
	if !hasExpression || !closed {
		kind = syntax.BogusExpression
	}
	return parenthesized.Complete(parser, kind)
}

func parseInlineMathAbsoluteExpression(parser *Parser) CompletedMarker {
	absolute := parser.Start()
	if !parser.EatWithContext(syntax.Pipe, lexer.LexInlineMath) {
		panic("grammar(script): expected opening absolute-value delimiter")
	}
	hasExpression := false
	if !parser.At(syntax.Pipe) {
		_, hasExpression = parseInlineMathExpression(parser, inlineMathLowestPrecedence)
	}
	closed := parser.EatWithContext(syntax.Pipe, lexer.LexInlineMath)
	kind := syntax.AbsoluteExpression
	if !hasExpression || !closed {
		kind = syntax.BogusExpression
	}
	return absolute.Complete(parser, kind)
}

func parseBogusInlineMathExpression(parser *Parser) {
	if isInlineMathEnd(parser) {
		return
	}
	bogus := parser.Start()
	for !isInlineMathEnd(parser) && !parser.HasPrecedingLineBreak() {
		parser.BumpWithContext(lexer.LexInlineMath)
	}
	bogus.Complete(parser, syntax.BogusExpression)
}

func inlineMathBinaryPrecedence(kind syntax.SyntaxKind) int {
	if kind == syntax.Plus || kind == syntax.Minus {
		return inlineMathAdditivePrecedence
	}
	if kind == syntax.Star || kind == syntax.Slash || kind == syntax.Percent {
		return inlineMathMultiplicativePrecedence
	}
	return 0
}

func isInlineMathEnd(parser *Parser) bool {
	return parser.At(syntax.RBracket) ||
		parser.At(syntax.RCurly) ||
		parser.At(syntax.EOF) ||
		parser.HasPrecedingLineBreak()
}
