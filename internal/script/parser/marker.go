package parser

import "github.com/arxbombus/jominia/internal/script/syntax"

type Marker struct {
	position int
}

type CompletedMarker struct {
	position int
}

func (m Marker) Complete(parser *Parser, kind syntax.SyntaxKind) CompletedMarker {
	event := &parser.events[m.position]
	if event.Type != EventStart || event.Kind != syntax.Tombstone {
		panic("parser(script): marker is not pointing at a tombstone")
	}
	event.Kind = kind
	parser.events = append(parser.events, Event{Type: EventFinish})
	return CompletedMarker(m)
}

// Precede starts a new node that becomes the completed node's parent. The event stream records a forward-parent link because the parent's start event is appended after its already-parsed left child.
func (m CompletedMarker) Precede(parser *Parser) Marker {
	event := &parser.events[m.position]
	if event.Type != EventStart || event.Kind == syntax.Tombstone {
		panic("parser(script): completed marker is not pointing at a node")
	}
	if event.ForwardParent != 0 {
		panic("parser(script): completed marker already has a forward parent")
	}
	parent := parser.Start()
	// Start may grow and reallocate the event slice, so look the child event up again instead of retaining a pointer across the append.
	parser.events[m.position].ForwardParent = parent.position - m.position
	return parent
}
