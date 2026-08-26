package parser

import (
	"github.com/arxbombus/jominia/internal/jomini/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

type EventType uint8

const (
	EventStart EventType = iota
	EventFinish
	EventToken
)

type Event struct {
	Type EventType
	Kind syntax.SyntaxKind
	End  text.TextSize
}

//nolint:unused // for the future :)
func tombstoneEvent() Event {
	return Event{
		Type: EventStart,
		Kind: syntax.Tombstone,
	}
}
