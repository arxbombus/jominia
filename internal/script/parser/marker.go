package parser

import "github.com/arxbombus/jominia/internal/script/syntax"

type Marker struct {
	position int
}

func (m Marker) Complete(parser *Parser, kind syntax.SyntaxKind) {
	event := &parser.events[m.position]
	if event.Type != EventStart || event.Kind != syntax.Tombstone {
		panic("parser: marker is not pointing at a tombstone")
	}
	event.Kind = kind
	parser.events = append(parser.events, Event{
		Type: EventFinish,
	})
}
