package parser

import "github.com/arxbombus/jominia/internal/tree"

// Parse parses source into a lossless green syntax tree.
func Parse(source string) *tree.GreenNode {
	parser := NewParser(source)
	parseRoot(parser)
	events, trivia := parser.Finish()
	sink := NewLosslessTreeSink(source, trivia)
	processEvents(sink, events)
	return sink.Finish()
}
