package parser

import (
	"strings"

	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
)

// parseScalar parses one logical scalar. Parameters, interpolated scalars, and variable references are re-lexed into source-derived parts; ordinary scalars remain single tokens.
func parseScalar(parser *Parser) bool {
	if parser.At(syntax.Dollar) || isStandaloneParameter(parser) {
		parseParameterExpression(parser)
		return true
	}
	if isInterpolatedIdentifier(parser) {
		parseInterpolatedIdentifier(parser)
		return true
	}
	if isInterpolatedString(parser) {
		parseInterpolatedString(parser)
		return true
	}
	if parser.At(syntax.At) || isVariableReference(parser) {
		parseVariableReference(parser)
		return true
	}
	if !parser.Current().IsScalar() {
		return false
	}
	parser.Bump()
	return true
}

// scanScalar returns the first token offset after one logical scalar.
//
// Normal-script lookahead sees structured scalars as one token until parsing contextually re-lexes them. The multi-token branches also make this helper correct if lookahead begins while a nested lexer mode is active.
func scanScalar(parser *Parser, offset int) (int, bool) {
	if parser.Nth(offset).IsScalar() {
		return offset + 1, true
	}
	if parser.Nth(offset) != syntax.Dollar {
		if parser.Nth(offset) != syntax.At {
			return offset, false
		}
		offset++
		if parser.Nth(offset) == syntax.Identifier {
			offset++
		}
		return offset, true
	}
	offset++
	if parser.Nth(offset) == syntax.ParameterName {
		offset++
	}
	if parser.Nth(offset) == syntax.Pipe {
		offset++
		if parser.Nth(offset) == syntax.ParameterArgument {
			offset++
		}
	}
	if parser.Nth(offset) == syntax.Dollar {
		offset++
	}
	return offset, true
}

func isStandaloneParameter(parser *Parser) bool {
	if !parser.Current().IsScalar() {
		return false
	}
	value := parser.CurrentText()
	if len(value) == 0 || value[0] != '$' {
		return false
	}
	offset := 1
	for offset < len(value) && isParameterNameByte(value[offset]) {
		offset++
	}
	if offset == len(value) {
		return true
	}
	if value[offset] == '$' {
		return offset == len(value)-1
	}
	if value[offset] != '|' {
		return false
	}
	closingOffset := strings.IndexByte(value[offset+1:], '$')
	return closingOffset < 0 || offset+closingOffset+1 == len(value)-1
}

func isInterpolatedIdentifier(parser *Parser) bool {
	return parser.At(syntax.Identifier) && hasParameterStart(parser.CurrentText(), 0, len(parser.CurrentText()), false)
}

func isInterpolatedString(parser *Parser) bool {
	if !parser.At(syntax.String) {
		return false
	}
	value := parser.CurrentText()
	return hasParameterStart(value, 1, len(value)-1, true) ||
		hasUnescapedBracketStart(value, 1, len(value)-1)
}

func hasUnescapedBracketStart(value string, start, end int) bool {
	for offset := start; offset < end; offset++ {
		if value[offset] == '\\' && offset+1 < end {
			offset++
			continue
		}
		if value[offset] == '[' {
			return true
		}
	}
	return false
}

func hasParameterStart(value string, start, end int, honorEscapes bool) bool {
	for offset := start; offset < end; offset++ {
		if honorEscapes && value[offset] == '\\' && offset+1 < end {
			offset++
			continue
		}
		if value[offset] != '$' {
			continue
		}
		next := offset + 1
		if next < end && (isParameterNameByte(value[next]) || value[next] == '|' || value[next] == '$') {
			return true
		}
	}
	return false
}

func isVariableReference(parser *Parser) bool {
	if !parser.At(syntax.Identifier) {
		return false
	}
	value := parser.CurrentText()
	return len(value) > 1 && value[0] == '@'
}

func isParameterNameByte(value byte) bool {
	return value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value == '_'
}

// parseParameterExpression is the canonical parser for parameters in both normal script and inline math.
func parseParameterExpression(parser *Parser) CompletedMarker {
	parameter := parser.Start()
	if !parser.At(syntax.Dollar) {
		if !isStandaloneParameter(parser) {
			panic("grammar(script): expected parameter opener")
		}
		parser.ReLex(lexer.ReLexParameter)
	}
	if !parser.Eat(syntax.Dollar) {
		panic("grammar(script): expected parameter opener")
	}
	hasName := parser.Eat(syntax.ParameterName)
	hasArgument := true
	if parser.Eat(syntax.Pipe) {
		hasArgument = parser.Eat(syntax.ParameterArgument)
	}
	closed := parser.Eat(syntax.Dollar)

	kind := syntax.ParameterExpression
	if !hasName || !hasArgument || !closed {
		kind = syntax.BogusExpression
	}
	return parameter.Complete(parser, kind)
}

func parseInterpolatedIdentifier(parser *Parser) CompletedMarker {
	if !isInterpolatedIdentifier(parser) {
		panic("grammar(script): expected interpolated identifier")
	}
	interpolated := parser.Start()
	end := parser.CurrentRange().End()
	parser.ReLex(lexer.ReLexInterpolatedIdentifier)
	for parser.CurrentRange().Start() < end {
		switch {
		case parser.At(syntax.IdentifierFragment):
			parser.Bump()
		case parser.At(syntax.Dollar):
			parseParameterExpression(parser)
		default:
			panic("grammar(script): unexpected token in interpolated identifier")
		}
	}
	return interpolated.Complete(parser, syntax.InterpolatedIdentifier)
}

func parseInterpolatedString(parser *Parser) CompletedMarker {
	if !isInterpolatedString(parser) {
		panic("grammar(script): expected interpolated string")
	}
	interpolated := parser.Start()
	parser.ReLex(lexer.ReLexInterpolatedString)
	if !parser.EatWithContext(syntax.StringQuote, lexer.LexInterpolatedString) {
		panic("grammar(script): expected interpolated string opener")
	}
	for !parser.At(syntax.StringQuote) && !parser.At(syntax.EOF) {
		switch {
		case parser.At(syntax.StringFragment):
			parser.BumpWithContext(lexer.LexInterpolatedString)
		case parser.At(syntax.Dollar):
			parseParameterExpression(parser)
		case parser.At(syntax.LBracket):
			parseBracketExpression(parser, lexer.LexInterpolatedString)
		default:
			bogus := parser.Start()
			parser.BumpWithContext(lexer.LexInterpolatedString)
			bogus.Complete(parser, syntax.BogusExpression)
		}
	}
	if !parser.EatWithContext(syntax.StringQuote, lexer.LexNormal) {
		panic("grammar(script): expected interpolated string closer")
	}
	return interpolated.Complete(parser, syntax.InterpolatedString)
}

func parseVariableReference(parser *Parser) CompletedMarker {
	variable := parser.Start()
	if !parser.At(syntax.At) {
		if !isVariableReference(parser) {
			panic("grammar(script): expected variable reference")
		}
		parser.ReLex(lexer.ReLexVariableReference)
	}
	if !parser.Eat(syntax.At) {
		panic("grammar(script): expected variable reference opener")
	}
	hasName := parser.Eat(syntax.Identifier)
	kind := syntax.VariableReference
	if !hasName {
		kind = syntax.BogusExpression
	}
	return variable.Complete(parser, kind)
}
