package parser

import (
	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
)

// parseBracketExpression parses the ordinary GUI-expression grammar. It is deliberately separate from arithmetic inline math: calls, member access, argument lists, and localization strings have their own syntax here.
func parseBracketExpression(parser *Parser, returnContext lexer.LexContext) CompletedMarker {
	bracket := parser.Start()
	if !parser.EatWithContext(syntax.LBracket, lexer.LexBracketExpression) {
		panic("grammar(script): expected bracket expression opener")
	}
	if !isBracketExpressionEnd(parser) {
		if _, ok := parseBracketOperand(parser); !ok {
			parseBogusBracketTokens(parser)
		}
		if parser.At(syntax.Pipe) {
			parseFormatSpecifier(parser)
		}
		if !isBracketExpressionEnd(parser) {
			parseBogusBracketTokens(parser)
		}
	}
	if parser.EatWithContext(syntax.RBracket, returnContext) {
		return bracket.Complete(parser, syntax.BracketExpression)
	}
	// The lexer stops an embedded malformed expression at the host string's quote. For a raw expression, restore normal tokenization at a recovery boundary without consuming the following statement.
	if returnContext == lexer.LexNormal && !parser.At(syntax.EOF) && !parser.At(syntax.StringQuote) {
		parser.ReLex(lexer.ReLexNormal)
	}
	return bracket.Complete(parser, syntax.BracketExpression)
}

func parseBracketOperand(parser *Parser) (CompletedMarker, bool) {
	if parser.At(syntax.Bang) || (parser.At(syntax.Question) && parser.Nth(1) != syntax.Dot) {
		unary := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		_, hasOperand := parseBracketOperand(parser)
		kind := syntax.UnaryExpression
		if !hasOperand {
			kind = syntax.BogusExpression
		}
		return unary.Complete(parser, kind), true
	}
	return parseBracketPostfixExpression(parser)
}

func parseBracketPostfixExpression(parser *Parser) (CompletedMarker, bool) {
	left, ok := parseBracketPrimaryExpression(parser)
	if !ok {
		return CompletedMarker{}, false
	}
	for {
		switch {
		case parser.At(syntax.LParen):
			call := left.Precede(parser)
			parseArgumentList(parser)
			left = call.Complete(parser, syntax.CallExpression)
		case parser.At(syntax.Dot), parser.At(syntax.Question) && parser.Nth(1) == syntax.Dot:
			member := left.Precede(parser)
			if parser.At(syntax.Question) {
				parser.BumpWithContext(lexer.LexBracketExpression)
			}
			parser.EatWithContext(syntax.Dot, lexer.LexBracketExpression)
			parseBracketMemberName(parser)
			left = member.Complete(parser, syntax.MemberExpression)
		default:
			return left, true
		}
	}
}

func parseBracketPrimaryExpression(parser *Parser) (CompletedMarker, bool) {
	switch {
	case parser.At(syntax.Number):
		number := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		return number.Complete(parser, syntax.NumberExpression), true
	case parser.At(syntax.Boolean):
		boolean := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		return boolean.Complete(parser, syntax.BooleanExpression), true
	case parser.At(syntax.Identifier):
		return parseBracketNameExpression(parser), true
	case parser.At(syntax.String), parser.At(syntax.SingleQuotedString):
		stringExpression := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		return stringExpression.Complete(parser, syntax.StringExpression), true
	case parser.At(syntax.Dollar):
		return parseParameterExpression(parser), true
	case parser.At(syntax.At):
		return parseVariableReference(parser), true
	case parser.At(syntax.LParen):
		return parseBracketParenthesizedExpression(parser), true
	case parser.At(syntax.LBracket):
		return parseBracketExpression(parser, lexer.LexBracketExpression), true
	case parser.At(syntax.ErrorToken):
		bogus := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		return bogus.Complete(parser, syntax.BogusExpression), true
	default:
		return CompletedMarker{}, false
	}
}

func parseBracketNameExpression(parser *Parser) CompletedMarker {
	name := parser.Start()
	parser.BumpWithContext(lexer.LexBracketExpression)
	return name.Complete(parser, syntax.NameExpression)
}

func parseBracketMemberName(parser *Parser) bool {
	if parser.At(syntax.Bang) || parser.At(syntax.Question) {
		unary := parser.Start()
		parser.BumpWithContext(lexer.LexBracketExpression)
		if !parser.At(syntax.Identifier) {
			unary.Complete(parser, syntax.BogusExpression)
			return false
		}
		parseBracketNameExpression(parser)
		unary.Complete(parser, syntax.UnaryExpression)
		return true
	}
	if !parser.At(syntax.Identifier) {
		return false
	}
	parseBracketNameExpression(parser)
	return true
}

func parseArgumentList(parser *Parser) CompletedMarker {
	arguments := parser.Start()
	if !parser.EatWithContext(syntax.LParen, lexer.LexBracketExpression) {
		panic("grammar(script): expected argument-list opener")
	}
	if parser.EatWithContext(syntax.RParen, lexer.LexBracketExpression) {
		return arguments.Complete(parser, syntax.ArgumentList)
	}

	for !isBracketArgumentListEnd(parser) {
		if _, ok := parseBracketOperand(parser); !ok {
			parseBogusBracketArgument(parser)
		}
		if parser.At(syntax.RParen) || isBracketExpressionEnd(parser) {
			break
		}
		if parser.At(syntax.Comma) {
			parser.BumpWithContext(lexer.LexBracketExpression)
			continue
		}

		// Preserve unexpected source between arguments inside the argument list.
		// This includes a would-be argument whose separating comma is missing.
		parseBogusBracketArgument(parser)
		if parser.At(syntax.Comma) {
			parser.BumpWithContext(lexer.LexBracketExpression)
			continue
		}
		break
	}
	parser.EatWithContext(syntax.RParen, lexer.LexBracketExpression)
	return arguments.Complete(parser, syntax.ArgumentList)
}

func parseBracketParenthesizedExpression(parser *Parser) CompletedMarker {
	parenthesized := parser.Start()
	parser.BumpWithContext(lexer.LexBracketExpression)
	_, hasExpression := parseBracketOperand(parser)
	closed := parser.EatWithContext(syntax.RParen, lexer.LexBracketExpression)
	kind := syntax.ParenthesizedExpression
	if !hasExpression || !closed {
		kind = syntax.BogusExpression
	}
	return parenthesized.Complete(parser, kind)
}

func parseFormatSpecifier(parser *Parser) CompletedMarker {
	format := parser.Start()
	parser.BumpWithContext(lexer.LexBracketExpression)
	parseBracketPrimaryExpression(parser)
	return format.Complete(parser, syntax.FormatSpecifier)
}

func parseBogusBracketArgument(parser *Parser) {
	if isBracketArgumentListEnd(parser) || parser.At(syntax.Comma) {
		return
	}
	bogus := parser.Start()
	for !isBracketArgumentListEnd(parser) && !parser.At(syntax.Comma) {
		parser.BumpWithContext(lexer.LexBracketExpression)
	}
	bogus.Complete(parser, syntax.BogusExpression)
}

func parseBogusBracketTokens(parser *Parser) {
	if isBracketExpressionEnd(parser) {
		return
	}
	bogus := parser.Start()
	for !isBracketExpressionEnd(parser) {
		parser.BumpWithContext(lexer.LexBracketExpression)
	}
	bogus.Complete(parser, syntax.BogusExpression)
}

func isBracketArgumentListEnd(parser *Parser) bool {
	return parser.At(syntax.RParen) || isBracketExpressionEnd(parser)
}

func isBracketExpressionEnd(parser *Parser) bool {
	if parser.At(syntax.RBracket) || parser.At(syntax.StringQuote) || parser.At(syntax.EOF) || parser.At(syntax.RCurly) {
		return true
	}
	return parser.HasPrecedingLineBreak() && isStrongBracketRecoveryStart(parser)
}

func isStrongBracketRecoveryStart(parser *Parser) bool {
	return parser.Current().IsScalar() && parser.Nth(1).IsOperator()
}
