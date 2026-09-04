package parser

import (
	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
)

type blockHeaderStatus uint8

const (
	blockHeaderNoMatch blockHeaderStatus = iota
	blockHeaderValid
	blockHeaderMalformed
)

type blockHeaderShape struct {
	leadingScalars    int
	hasOperator       bool
	hasTrailingScalar bool
}

type blockHeaderScan struct {
	status      blockHeaderStatus
	shape       blockHeaderShape
	blockOffset int
}

// parseRoot parses the source file into a root node.
func parseRoot(parser *Parser) {
	root := parser.Start()
	parseStatementList(parser, syntax.EOF)
	root.Complete(parser, syntax.Root)
}

// parseStatementList parses statements until endKind or EOF is reached.
func parseStatementList(parser *Parser, endKind syntax.SyntaxKind) {
	list := parser.Start()
	for !parser.At(syntax.EOF) && !parser.At(endKind) {
		if endKind != syntax.EOF && isClosingDelimiter(parser.Current()) {
			break
		}
		if parseStatement(parser) {
			continue
		}
		recoverStatement(parser, endKind)
	}
	list.Complete(parser, syntax.StatementList)
}

// parseStatement parses a single statement and reports whether parsing succeeded.
func parseStatement(parser *Parser) bool {
	switch {
	case isConditionalBlockStart(parser):
		parseConditionalBlock(parser)
		return true
	case parser.At(syntax.LCurly), parser.At(syntax.LBracket):
		parseValueStatement(parser)
		return true
	case parser.At(syntax.InlineMathStart):
		parseValueStatement(parser)
		return true
	case parser.Current().IsScalar():
		scan := scanBlockHeader(parser)
		switch scan.status {
		case blockHeaderValid:
			parseBlockStatement(parser, scan.shape)
		case blockHeaderMalformed:
			parseBogusBlockStatement(parser, scan.blockOffset)
		case blockHeaderNoMatch:
			if parser.Nth(1).IsOperator() {
				if parser.Nth(2).IsOperator() && !parser.NthHasPrecedingLineBreak(2) {
					parseBogusBinaryStatement(parser)
				} else {
					parseBinaryStatement(parser)
				}
			} else {
				parseValueStatement(parser)
			}
		default:
			panic("grammar(script): unknown block header status")
		}
		return true
	default:
		return false
	}
}

// parseValueStatement parses a positional value appearing in statement position.
func parseValueStatement(parser *Parser) {
	statement := parser.Start()
	if !parseValue(parser) {
		panic("grammar(script): expected value statement")
	}
	parser.Eat(syntax.Semicolon)
	statement.Complete(parser, syntax.ValueStatement)
}

// parseBinaryStatement parses a scalar, operator, and unbraced value list.
func parseBinaryStatement(parser *Parser) {
	statement := parser.Start()
	if !parseScalar(parser) {
		panic("grammar(script): binary statement must begin with a scalar")
	}
	if !parser.Current().IsOperator() {
		panic("grammar(script): binary statement is missing an operator")
	}
	parser.Bump()
	parseValueList(parser)
	parser.Eat(syntax.Semicolon)
	statement.Complete(parser, syntax.BinaryStatement)
}

// parseBlockStatement parses a block and the syntactic header immediately preceding it.
func parseBlockStatement(parser *Parser, shape blockHeaderShape) {
	statement := parser.Start()
	parseBlockHeader(parser, shape)
	if !parser.At(syntax.LCurly) {
		panic("grammar(script): block header is not followed by a block")
	}
	parseBlock(parser)
	parser.Eat(syntax.Semicolon)
	statement.Complete(parser, syntax.BlockStatement)
}

// parseBlockHeader parses the scalar and operator sequence preceding a block.
func parseBlockHeader(parser *Parser, shape blockHeaderShape) {
	header := parser.Start()
	parseScalarList(parser, shape.leadingScalars)
	if shape.hasOperator {
		if !parser.Current().IsOperator() {
			panic("grammar(script): expected block header operator")
		}
		parser.Bump()
	}
	if shape.hasTrailingScalar {
		parseScalarList(parser, 1)
	}
	header.Complete(parser, syntax.BlockHeader)
}

// parseScalarList parses exactly count adjacent scalars.
func parseScalarList(parser *Parser, count int) {
	if count <= 0 {
		panic("grammar(script): scalar list must contain at least one scalar")
	}
	list := parser.Start()
	for i := 0; i < count; i++ {
		if i > 0 && parser.HasPrecedingLineBreak() {
			panic("grammar(script): scalar list cannot cross a line break")
		}
		if !parseScalar(parser) {
			panic("grammar(script): expected scalar in scalar list")
		}
	}
	list.Complete(parser, syntax.ScalarList)
}

// parseValueList parses the unbraced right-hand-side values of a binary statement.
func parseValueList(parser *Parser) {
	list := parser.Start()
	if parseValue(parser) {
		for !isValueListTerminator(parser.Current()) {
			if parser.HasPrecedingLineBreak() {
				break
			}
			if isConditionalBlockStart(parser) {
				break
			}
			if isBlockLikeStatementStart(parser) {
				break
			}
			if parser.Current().IsScalar() && parser.Nth(1).IsOperator() {
				break
			}
			if !parseValue(parser) {
				break
			}
		}
	}
	list.Complete(parser, syntax.ValueList)
}

// parseValue parses one syntactic value.
func parseValue(parser *Parser) bool {
	switch {
	case parseScalar(parser):
		if parser.At(syntax.LBracket) && !parser.HasPrecedingLineBreak() {
			parseBracketExpression(parser, lexer.LexNormal)
		}
		return true
	case parser.At(syntax.LCurly):
		parseBlock(parser)
		return true
	case parser.At(syntax.LBracket):
		parseBracketExpression(parser, lexer.LexNormal)
		return true
	case parser.At(syntax.InlineMathStart):
		parseInlineMath(parser)
		return true
	case parser.At(syntax.ErrorToken):
		bogus := parser.Start()
		parser.Bump()
		bogus.Complete(parser, syntax.BogusExpression)
		return true
	default:
		return false
	}
}

// parseBlock parses a brace-delimited statement list.
func parseBlock(parser *Parser) {
	block := parser.Start()
	if !parser.Eat(syntax.LCurly) {
		panic("grammar(script): expected opening block delimiter")
	}
	parseStatementList(parser, syntax.RCurly)
	parser.Eat(syntax.RCurly)
	block.Complete(parser, syntax.Block)
}

// scanBlockHeader classifies the scalar sequence at the current token as a valid block header, a malformed block-like header, or not a block header.
//
// The generic forms supported here are:
//
//	foo {
//	types wargoal_types {
//	foo = {
//	color = rgb {
//	type add_wargoal_panel = default_block_window {
//	blockoverride "name" {
func scanBlockHeader(parser *Parser) blockHeaderScan {
	offset, ok := scanScalar(parser, 0)
	if !ok {
		return blockHeaderScan{status: blockHeaderNoMatch}
	}
	shape := blockHeaderShape{leadingScalars: 1}
	if nextOffset, hasScalar := scanScalar(parser, offset); hasScalar && !parser.NthHasPrecedingLineBreak(offset) {
		shape.leadingScalars = 2
		offset = nextOffset
	}
	if parser.Nth(offset) == syntax.LCurly {
		return blockHeaderScan{
			status:      blockHeaderValid,
			shape:       shape,
			blockOffset: offset,
		}
	}
	if !parser.Nth(offset).IsOperator() {
		return blockHeaderScan{status: blockHeaderNoMatch}
	}
	shape.hasOperator = true
	offset++
	if parser.Nth(offset) == syntax.LCurly {
		return blockHeaderScan{
			status:      blockHeaderValid,
			shape:       shape,
			blockOffset: offset,
		}
	}
	if parser.Nth(offset).IsOperator() {
		if blockOffset, ok := malformedBlockOffset(parser, offset); ok {
			return blockHeaderScan{
				status:      blockHeaderMalformed,
				blockOffset: blockOffset,
			}
		}
		return blockHeaderScan{status: blockHeaderNoMatch}
	}
	nextOffset, hasTrailingScalar := scanScalar(parser, offset)
	if !hasTrailingScalar {
		return blockHeaderScan{status: blockHeaderNoMatch}
	}
	shape.hasTrailingScalar = true
	offset = nextOffset
	if parser.Nth(offset) == syntax.LCurly {
		return blockHeaderScan{
			status:      blockHeaderValid,
			shape:       shape,
			blockOffset: offset,
		}
	}
	if parser.Nth(offset).IsOperator() {
		if blockOffset, ok := malformedBlockOffset(parser, offset); ok {
			return blockHeaderScan{
				status:      blockHeaderMalformed,
				blockOffset: blockOffset,
			}
		}
	}
	return blockHeaderScan{status: blockHeaderNoMatch}
}

// malformedBlockOffset reports the opening block after an extra operator in an otherwise block-like header. extraOperatorOffset points at the unexpected operator.
func malformedBlockOffset(parser *Parser, extraOperatorOffset int) (int, bool) {
	nextOffset := extraOperatorOffset + 1
	if parser.Nth(nextOffset) == syntax.LCurly {
		return nextOffset, true
	}
	if afterScalar, hasScalar := scanScalar(parser, nextOffset); hasScalar && parser.Nth(afterScalar) == syntax.LCurly {
		return afterScalar, true
	}
	return 0, false
}

// parseBogusBlockStatement preserves an obviously block-like malformed statement as one recovery node while retaining its nested block structure.
func parseBogusBlockStatement(parser *Parser, blockOffset int) {
	bogus := parser.Start()
	for i := 0; i < blockOffset; i++ {
		if !parseScalar(parser) {
			parser.Bump()
		}
	}
	if !parser.At(syntax.LCurly) {
		panic("grammar(script): malformed block statement is missing its block")
	}
	parseBlock(parser)
	parser.Eat(syntax.Semicolon)
	bogus.Complete(parser, syntax.BogusStatement)
}

// parseBogusBinaryStatement recovers a same-line binary statement containing a repeated operator while preserving following recognizable statements.
func parseBogusBinaryStatement(parser *Parser) {
	bogus := parser.Start()
	if !parseScalar(parser) {
		panic("grammar(script): bogus binary statement must begin with a scalar")
	}
	parser.Bump() // first operator
	for !parser.At(syntax.EOF) {
		if parser.At(syntax.Semicolon) {
			parser.Bump()
			break
		}
		if isClosingDelimiter(parser.Current()) {
			break
		}
		if parser.HasPrecedingLineBreak() {
			break
		}
		if isStrongStatementStart(parser) {
			break
		}
		parser.Bump()
	}
	bogus.Complete(parser, syntax.BogusStatement)
}

// recoverStatement consumes invalid input until parsing can resume at a reliable statement boundary.
func recoverStatement(parser *Parser, endKind syntax.SyntaxKind) {
	if parser.At(syntax.EOF) || parser.At(endKind) {
		return
	}
	bogus := parser.Start()
	// Always consume at least one token so recovery makes progress.
	parser.Bump()
	for !parser.At(syntax.EOF) && !parser.At(endKind) {
		if parser.At(syntax.Semicolon) {
			parser.Bump()
			break
		}
		if endKind != syntax.EOF && isClosingDelimiter(parser.Current()) {
			break
		}
		if parser.HasPrecedingLineBreak() {
			break
		}
		if isStrongStatementStart(parser) {
			break
		}
		parser.Bump()
	}
	bogus.Complete(parser, syntax.BogusStatement)
}

// isStrongStatementStart reports whether the current token begins a statement with enough structure to be a useful recovery boundary.
func isStrongStatementStart(parser *Parser) bool {
	if parser.At(syntax.LCurly) ||
		parser.At(syntax.LBracket) ||
		parser.At(syntax.InlineMathStart) {
		return true
	}
	if !parser.Current().IsScalar() {
		return false
	}
	if scanBlockHeader(parser).status != blockHeaderNoMatch {
		return true
	}
	return parser.Nth(1).IsOperator()
}

// isBlockLikeStatementStart reports whether the current token begins either a valid or recognizably malformed block statement.
func isBlockLikeStatementStart(parser *Parser) bool {
	return scanBlockHeader(parser).status != blockHeaderNoMatch
}

// isValueListTerminator reports whether kind terminates an unbraced value list.
func isValueListTerminator(kind syntax.SyntaxKind) bool {
	return kind == syntax.EOF ||
		kind == syntax.RCurly ||
		kind == syntax.RBracket ||
		kind == syntax.RParen ||
		kind == syntax.Semicolon
}

// isClosingDelimiter reports whether kind closes a delimited construct.
func isClosingDelimiter(kind syntax.SyntaxKind) bool {
	return kind == syntax.RCurly ||
		kind == syntax.RBracket ||
		kind == syntax.RParen
}

// isConditionalBlockStart recognizes the adjacent [[ opener without changing how ordinary, possibly nested bracket groups are tokenized.
func isConditionalBlockStart(parser *Parser) bool {
	return parser.At(syntax.LBracket) &&
		parser.Nth(1) == syntax.LBracket &&
		parser.CurrentRange().End() == parser.NthRange(1).Start()
}

// parseConditionalBlock parses [[CONDITION] statements...]. The outer pair of brackets bounds the block; ConditionalHeader owns the nested [CONDITION] pair. The body deliberately reuses StatementList so nested conditionals and all ordinary statements share the same recovery behavior.
func parseConditionalBlock(parser *Parser) {
	conditional := parser.Start()
	if !parser.Eat(syntax.LBracket) {
		panic("grammar(script): expected conditional block opener")
	}
	parseConditionalHeader(parser)
	parseStatementList(parser, syntax.RBracket)
	parser.Eat(syntax.RBracket)
	conditional.Complete(parser, syntax.ConditionalBlock)
}

func parseConditionalHeader(parser *Parser) {
	header := parser.Start()
	if !parser.Eat(syntax.LBracket) {
		panic("grammar(script): expected conditional header opener")
	}
	parser.Eat(syntax.Bang)
	parseScalar(parser)
	parser.Eat(syntax.RBracket)
	header.Complete(parser, syntax.ConditionalHeader)
}
