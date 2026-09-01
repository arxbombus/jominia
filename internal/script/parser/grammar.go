package parser

import "github.com/arxbombus/jominia/internal/script/syntax"

// parseRoot parses the source file into a root node.
func parseRoot(parser *Parser) {
	root := parser.Start()
	parseEntryList(parser, syntax.EOF)
	root.Complete(parser, syntax.Root)
}

// parseEntryList parses entries until endKind or EOF is reached.
func parseEntryList(parser *Parser, endKind syntax.SyntaxKind) {
	for !parser.At(syntax.EOF) && !parser.At(endKind) {
		if endKind != syntax.EOF && isClosingDelimiter(parser.Current()) {
			return
		}
		if parseEntry(parser) {
			continue
		}
		recoverEntry(parser, endKind)
	}
}

// parseEntry parses a single entry and reports whether parsing succeeded.
func parseEntry(parser *Parser) bool {
	switch {
	case parser.At(syntax.LCurly):
		entry := parser.Start()
		parseBlock(parser)
		parser.Eat(syntax.Semicolon)
		entry.Complete(parser, syntax.Entry)
		return true
	case parser.At(syntax.LBracket):
		entry := parser.Start()
		parseBracketGroup(parser)
		parser.Eat(syntax.Semicolon)
		entry.Complete(parser, syntax.Entry)
		return true
	case parser.Current().IsScalar():
		entry := parser.Start()
		parser.Bump()
		parseScalarEntryTail(parser)
		parser.Eat(syntax.Semicolon)
		entry.Complete(parser, syntax.Entry)
		return true
	default:
		return false
	}
}

// parseScalarEntryTail parses the portion of an entry following its leading scalar.
func parseScalarEntryTail(parser *Parser) {
	if parser.Current().IsOperator() {
		parser.Bump()
		parseValueSequence(parser)
		return
	}
	if parser.At(syntax.LCurly) {
		parseBlock(parser)
	}
}

// parseValueSequence parses one or more values belonging to the same entry.
func parseValueSequence(parser *Parser) {
	if !parseValue(parser) {
		return
	}
	for {
		//exhaustive:ignore
		switch parser.Current() {
		case syntax.EOF,
			syntax.RCurly,
			syntax.RBracket,
			syntax.RParen,
			syntax.Semicolon:
			return
		}
		if parser.HasPrecedingLineBreak() {
			return
		}
		if parser.Current().IsScalar() && (parser.Nth(1).IsOperator() || parser.Nth(1) == syntax.LCurly) {
			return
		}
		if !parseValue(parser) {
			return
		}
	}
}

// parseValue parses a scalar, block, or bracket group value.
func parseValue(parser *Parser) bool {
	switch {
	case parser.Current().IsScalar():
		parser.Bump()
		if parser.At(syntax.LCurly) {
			parseBlock(parser)
			return true
		}
		if parser.At(syntax.LBracket) && !parser.HasPrecedingLineBreak() {
			parseBracketGroup(parser)
		}
		return true
	case parser.At(syntax.LCurly):
		parseBlock(parser)
		return true
	case parser.At(syntax.LBracket):
		parseBracketGroup(parser)
		return true
	default:
		return false
	}
}

// parseBlock parses a brace-delimited block.
func parseBlock(parser *Parser) {
	block := parser.Start()
	parser.Bump()
	parseEntryList(parser, syntax.RCurly)
	parser.Eat(syntax.RCurly)
	block.Complete(parser, syntax.Block)
}

// parseBracketGroup parses a bracket-delimited opaque group.
func parseBracketGroup(parser *Parser) {
	parseOpaque(parser, syntax.LBracket, syntax.RBracket, syntax.BracketGroup)
}

// parseParenGroup parses a parenthesis-delimited opaque group.
func parseParenGroup(parser *Parser) {
	parseOpaque(parser, syntax.LParen, syntax.RParen, syntax.ParenGroup)
}

// parseOpaque parses a delimited group without interpreting its contents, while still preserving nested blocks and groups structurally.
func parseOpaque(parser *Parser, startKind, endKind, nodeKind syntax.SyntaxKind) {
	opaque := parser.Start()
	if !parser.Eat(startKind) {
		panic("grammar: expected start delimiter not found")
	}
	for !parser.At(endKind) && !parser.At(syntax.EOF) {
		if isClosingDelimiter(parser.Current()) {
			break
		}
		switch {
		case parser.At(syntax.LCurly):
			parseBlock(parser)
		case parser.At(syntax.LBracket):
			parseBracketGroup(parser)
		case parser.At(syntax.LParen):
			parseParenGroup(parser)
		default:
			parser.Bump()
		}
	}
	parser.Eat(endKind)
	opaque.Complete(parser, nodeKind)
}

// recoverEntry consumes invalid input until parsing can resume at another entry.
func recoverEntry(parser *Parser, endKind syntax.SyntaxKind) {
	if parser.At(syntax.EOF) || parser.At(endKind) {
		return
	}
	bogus := parser.Start()
	// Always consume at least one token so recovery makes progress.
	parser.Bump()
	for !parser.At(syntax.EOF) && !parser.At(endKind) {
		if endKind != syntax.EOF && isClosingDelimiter(parser.Current()) {
			break
		}
		if parser.HasPrecedingLineBreak() {
			break
		}
		if isEntryStart(parser.Current()) {
			break
		}
		parser.Bump()
	}
	bogus.Complete(parser, syntax.Bogus)
}

// isEntryStart reports whether kind can begin an entry.
func isEntryStart(kind syntax.SyntaxKind) bool {
	return kind.IsScalar() ||
		kind == syntax.LCurly ||
		kind == syntax.LBracket
}

// isClosingDelimiter reports whether kind closes a delimited construct.
func isClosingDelimiter(kind syntax.SyntaxKind) bool {
	return kind == syntax.RCurly ||
		kind == syntax.RBracket ||
		kind == syntax.RParen
}
