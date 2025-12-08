package event_test

import (
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEvent struct {
	event.Base
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

func (e TestEvent) Payload() any {
	return e
}

func TestNewBase(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		aggregateType string
		aggregateID   string
	}{
		{
			name:          "user created event",
			eventType:     "user.created",
			aggregateType: "user",
			aggregateID:   "123",
		},
		{
			name:          "order placed event",
			eventType:     "order.placed",
			aggregateType: "order",
			aggregateID:   "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UTC()
			base := event.NewBase(tt.eventType, tt.aggregateType, tt.aggregateID)
			after := time.Now().UTC()

			assert.NotEqual(t, uuid.Nil, base.EventID())
			assert.Equal(t, tt.eventType, base.EventType())
			assert.Equal(t, tt.aggregateType, base.AggregateType())
			assert.Equal(t, tt.aggregateID, base.AggregateID())
			assert.True(t, base.OccurredAt().After(before) || base.OccurredAt().Equal(before))
			assert.True(t, base.OccurredAt().Before(after) || base.OccurredAt().Equal(after))
		})
	}
}

func TestRaisesEvents(t *testing.T) {
	t.Run("raise and get events", func(t *testing.T) {
		var r event.RaisesEvents

		e1 := TestEvent{
			Base:   event.NewBase("user.created", "user", "1"),
			UserID: uuid.New(),
			Email:  "test@example.com",
		}
		e2 := TestEvent{
			Base:   event.NewBase("user.updated", "user", "1"),
			UserID: uuid.New(),
			Email:  "updated@example.com",
		}

		r.Raise(e1)
		r.Raise(e2)

		events := r.Events()
		require.Len(t, events, 2)
		assert.Equal(t, "user.created", events[0].EventType())
		assert.Equal(t, "user.updated", events[1].EventType())
	})

	t.Run("clear events", func(t *testing.T) {
		var r event.RaisesEvents

		r.Raise(TestEvent{
			Base: event.NewBase("user.created", "user", "1"),
		})

		require.Len(t, r.Events(), 1)

		r.ClearEvents()
		assert.Empty(t, r.Events())
	})

	t.Run("empty events", func(t *testing.T) {
		var r event.RaisesEvents
		assert.Empty(t, r.Events())
	})
}

func TestNewOutboxEvent(t *testing.T) {
	e := TestEvent{
		Base:   event.NewBase("user.created", "user", "123"),
		UserID: uuid.New(),
		Email:  "test@example.com",
	}
	payload := []byte(`{"user_id":"123","email":"test@example.com"}`)

	outbox := event.NewOutboxEvent(e, payload)

	assert.Equal(t, e.EventID(), outbox.ID)
	assert.Equal(t, "user", outbox.AggregateType)
	assert.Equal(t, "123", outbox.AggregateID)
	assert.Equal(t, "user.created", outbox.EventType)
	assert.Equal(t, payload, outbox.Payload)
	assert.Equal(t, e.OccurredAt(), outbox.CreatedAt)
	assert.Nil(t, outbox.PublishedAt)
	assert.Equal(t, 0, outbox.RetryCount)
	assert.Nil(t, outbox.LastError)
}
