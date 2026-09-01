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
	Type EventType
	Kind syntax.SyntaxKind
	End  text.TextSize
}

// processEvents processes a list of parser events and applies them to the given tree sink.
func processEvents(sink TreeSink, events []Event) {
	for _, event := range events {
		switch event.Type {
		case EventStart:
			if event.Kind == syntax.Tombstone {
				continue
			}
			sink.StartNode(event.Kind)
		case EventFinish:
			sink.FinishNode()
		case EventToken:
			sink.Token(event.Kind, event.End)
		default:
			panic("parser: unknown event type")
		}
	}
}

func tombstoneEvent() Event {
	return Event{
		Type: EventStart,
		Kind: syntax.Tombstone,
	}
}
