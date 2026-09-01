package parser

import (
	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

type EventType uint8

const (
	EventStart EventType = iota
	EventFinish
	EventToken
)

// Event represents a single parser event, such as starting a node, finishing a node, or encountering a token.
type Event struct {
	Type          EventType
	Kind          syntax.SyntaxKind
	End           text.TextSize
	ForwardParent int
}

// processEvents processes a list of parser events and applies them to the given tree sink.
func processEvents(sink TreeSink, events []Event) {
	for position := 0; position < len(events); position++ {
		event := events[position]
		switch event.Type {
		case EventStart:
			if event.Kind == syntax.Tombstone {
				continue
			}
			kinds := []syntax.SyntaxKind{event.Kind}
			parentPosition := position
			forwardParent := event.ForwardParent
			for forwardParent != 0 {
				nextParentPosition := parentPosition + forwardParent
				if nextParentPosition <= parentPosition || nextParentPosition >= len(events) {
					panic("parser(script): invalid forward parent")
				}
				parentPosition = nextParentPosition
				parent := events[parentPosition]
				if parent.Type != EventStart || parent.Kind == syntax.Tombstone {
					panic("parser(script): forward parent is not a start event")
				}
				kinds = append(kinds, parent.Kind)
				forwardParent = parent.ForwardParent
				events[parentPosition] = tombstoneEvent()
			}
			for index := len(kinds) - 1; index >= 0; index-- {
				sink.StartNode(kinds[index])
			}
		case EventFinish:
			sink.FinishNode()
		case EventToken:
			sink.Token(event.Kind, event.End)
		default:
			panic("parser(script): unknown event type")
		}
	}
}

func tombstoneEvent() Event {
	return Event{Type: EventStart, Kind: syntax.Tombstone}
}
