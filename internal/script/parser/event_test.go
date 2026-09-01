package parser

import (
	"testing"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

type treeSinkCall struct {
	eventType EventType
	kind      syntax.SyntaxKind
	end       text.TextSize
}

type recordingTreeSink struct {
	calls []treeSinkCall
}

func (s *recordingTreeSink) Token(
	kind syntax.SyntaxKind,
	end text.TextSize,
) {
	s.calls = append(s.calls, treeSinkCall{
		eventType: EventToken,
		kind:      kind,
		end:       end,
	})
}

func (s *recordingTreeSink) StartNode(kind syntax.SyntaxKind) {
	s.calls = append(s.calls, treeSinkCall{
		eventType: EventStart,
		kind:      kind,
	})
}

func (s *recordingTreeSink) FinishNode() {
	s.calls = append(s.calls, treeSinkCall{
		eventType: EventFinish,
	})
}

func TestTombstoneEvent(t *testing.T) {
	event := tombstoneEvent()

	if event.Type != EventStart {
		t.Fatalf("expected EventStart, got %d", event.Type)
	}

	if event.Kind != syntax.Tombstone {
		t.Fatalf("expected Tombstone, got %s", event.Kind)
	}
}

func TestProcessEventsDispatchesEvents(t *testing.T) {
	events := []Event{
		{
			Type: EventStart,
			Kind: syntax.Root,
		},
		{
			Type: EventStart,
			Kind: syntax.Entry,
		},
		{
			Type: EventToken,
			Kind: syntax.Identifier,
			End:  3,
		},
		{
			Type: EventToken,
			Kind: syntax.Equals,
			End:  5,
		},
		{
			Type: EventToken,
			Kind: syntax.Identifier,
			End:  9,
		},
		{
			Type: EventFinish,
		},
		{
			Type: EventFinish,
		},
	}

	sink := &recordingTreeSink{}
	processEvents(sink, events)

	expected := []treeSinkCall{
		{
			eventType: EventStart,
			kind:      syntax.Root,
		},
		{
			eventType: EventStart,
			kind:      syntax.Entry,
		},
		{
			eventType: EventToken,
			kind:      syntax.Identifier,
			end:       3,
		},
		{
			eventType: EventToken,
			kind:      syntax.Equals,
			end:       5,
		},
		{
			eventType: EventToken,
			kind:      syntax.Identifier,
			end:       9,
		},
		{
			eventType: EventFinish,
		},
		{
			eventType: EventFinish,
		},
	}

	if len(sink.calls) != len(expected) {
		t.Fatalf(
			"expected %d sink calls, got %d",
			len(expected),
			len(sink.calls),
		)
	}

	for i, want := range expected {
		got := sink.calls[i]

		if got != want {
			t.Errorf(
				"call %d = %+v, want %+v",
				i,
				got,
				want,
			)
		}
	}
}

func TestProcessEventsSkipsTombstones(t *testing.T) {
	events := []Event{
		tombstoneEvent(),
		{
			Type: EventStart,
			Kind: syntax.Root,
		},
		{
			Type: EventFinish,
		},
	}

	sink := &recordingTreeSink{}
	processEvents(sink, events)

	if len(sink.calls) != 2 {
		t.Fatalf("expected 2 sink calls, got %d", len(sink.calls))
	}

	if sink.calls[0].eventType != EventStart ||
		sink.calls[0].kind != syntax.Root {
		t.Fatalf("unexpected first sink call: %+v", sink.calls[0])
	}

	if sink.calls[1].eventType != EventFinish {
		t.Fatalf("unexpected second sink call: %+v", sink.calls[1])
	}
}

func TestProcessEventsRejectsUnknownEventType(t *testing.T) {
	events := []Event{
		{
			Type: EventType(255),
		},
	}

	sink := &recordingTreeSink{}

	defer func() {
		if recover() == nil {
			t.Fatal("expected unknown event type to panic")
		}
	}()

	processEvents(sink, events)
}
