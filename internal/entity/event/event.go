package event

import (
	"time"

	"github.com/google/uuid"
)

type Event interface {
	EventID() uuid.UUID
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
	Payload() any
}

type Base struct {
	ID        uuid.UUID
	Type      string
	AggrID    string
	AggrType  string
	Timestamp time.Time
}

func NewBase(eventType, aggregateType, aggregateID string) Base {
	return Base{
		ID:        uuid.New(),
		Type:      eventType,
		AggrType:  aggregateType,
		AggrID:    aggregateID,
		Timestamp: time.Now().UTC(),
	}
}

func (b *Base) EventID() uuid.UUID    { return b.ID }
func (b *Base) EventType() string     { return b.Type }
func (b *Base) AggregateID() string   { return b.AggrID }
func (b *Base) AggregateType() string { return b.AggrType }
func (b *Base) OccurredAt() time.Time { return b.Timestamp }

type Aggregate interface {
	Events() []Event
	ClearEvents()
}

type RaisesEvents struct {
	events []Event
}

func (r *RaisesEvents) Raise(e Event) {
	r.events = append(r.events, e)
}

func (r *RaisesEvents) Events() []Event {
	return r.events
}

func (r *RaisesEvents) ClearEvents() {
	r.events = nil
}
