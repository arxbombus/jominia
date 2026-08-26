package parser

import "github.com/arxbombus/jominia/internal/jomini/syntax"

func parseRoot(parser *Parser) {
	root := parser.Start()
	for !parser.At(syntax.EOF) {
		parseEntry(parser)
	}
	root.Complete(parser, syntax.Root)
}

func parseEntry(parser *Parser) {
	entry := parser.Start()
	parser.Bump()
	/*
		foo = { bar = baz }
		foo { bar = baz }
	*/
	if parser.Eat(syntax.Equals) {
		parseValue(parser)
	} else if parser.At(syntax.LCurly) {
		parseBlock(parser)
	}
	entry.Complete(parser, syntax.Entry)
}

func parseValue(parser *Parser) {
	if parser.At(syntax.LCurly) {
		parseBlock(parser)
		return
	}
	parser.Bump()
}

func parseBlock(parser *Parser) {
	block := parser.Start()
	parser.Bump()
	for !parser.At(syntax.RCurly) && !parser.At(syntax.EOF) {
		parseEntry(parser)
	}
	parser.Eat(syntax.RCurly)
	block.Complete(parser, syntax.Block)
}
