package event

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
	PublishedAt   *time.Time
	RetryCount    int
	LastError     *string
}

func NewOutboxEvent(e Event, payload []byte) OutboxEvent {
	return OutboxEvent{
		ID:            e.EventID(),
		AggregateType: e.AggregateType(),
		AggregateID:   e.AggregateID(),
		EventType:     e.EventType(),
		Payload:       payload,
		CreatedAt:     e.OccurredAt(),
		RetryCount:    0,
	}
}
