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

// parseInlineMath parses an @[...] (or @\[...]) expression value. Only the current token stream is re-lexed: the resulting expression tokens flow through the same event stream and lossless green-tree sink as normal script.
func parseInlineMath(parser *Parser) {
	inlineMath := parser.Start()
	if !parser.Eat(syntax.InlineMathStart) {
		panic("grammar(script): expected inline math opener")
	}

	parser.ReLex(lexer.ReLexInlineMath)
	if !isInlineMathEnd(parser) {
		if _, ok := parseInlineMathExpression(parser, inlineMathLowestPrecedence); !ok {
			parseBogusInlineMathExpression(parser)
		}
	}

	if !isInlineMathEnd(parser) {
		parseBogusInlineMathExpression(parser)
	}

	if !parser.Eat(syntax.RBracket) && !parser.At(syntax.EOF) {
		// The inline lexer normally returns to normal mode when it emits ]. If
		// recovery stopped at another boundary, restore normal tokenization for
		// the unconsumed token and the statements that follow it.
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
		parser.Bump()
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
		parser.Bump()
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
		parser.Bump()
		return number.Complete(parser, syntax.NumberExpression), true
	case parser.At(syntax.Identifier), parser.At(syntax.At):
		return parseInlineMathNameExpression(parser), true
	case parser.At(syntax.Dollar):
		return parseInlineMathParameterExpression(parser), true
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
	hasAt := parser.Eat(syntax.At)
	hasName := parser.Eat(syntax.Identifier)
	kind := syntax.NameExpression
	if hasAt && !hasName {
		kind = syntax.BogusExpression
	}
	return name.Complete(parser, kind)
}

func parseInlineMathParameterExpression(parser *Parser) CompletedMarker {
	parameter := parser.Start()
	if !parser.Eat(syntax.Dollar) {
		panic("grammar(script): expected inline math parameter opener")
	}
	hasName := parser.Eat(syntax.ParameterName)
	if parser.Eat(syntax.Pipe) {
		parser.Eat(syntax.ParameterArgument)
	}
	closed := parser.Eat(syntax.Dollar)
	kind := syntax.ParameterExpression
	if !hasName || !closed {
		kind = syntax.BogusExpression
	}
	return parameter.Complete(parser, kind)
}

func parseInlineMathParenthesizedExpression(parser *Parser) CompletedMarker {
	parenthesized := parser.Start()
	if !parser.Eat(syntax.LParen) {
		panic("grammar(script): expected opening parenthesis")
	}
	_, hasExpression := parseInlineMathExpression(parser, inlineMathLowestPrecedence)
	closed := parser.Eat(syntax.RParen)
	kind := syntax.ParenthesizedExpression
	if !hasExpression || !closed {
		kind = syntax.BogusExpression
	}
	return parenthesized.Complete(parser, kind)
}

func parseInlineMathAbsoluteExpression(parser *Parser) CompletedMarker {
	absolute := parser.Start()
	if !parser.Eat(syntax.Pipe) {
		panic("grammar(script): expected opening absolute-value delimiter")
	}

	hasExpression := false
	if !parser.At(syntax.Pipe) {
		_, hasExpression = parseInlineMathExpression(parser, inlineMathLowestPrecedence)
	}
	closed := parser.Eat(syntax.Pipe)
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
		parser.Bump()
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
