package eventbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRabbitMQSubscriber_NewRabbitMQSubscriber_InvalidURL(t *testing.T) {
	t.Parallel()

	subscriber, err := NewRabbitMQSubscriber("invalid-url", "test-exchange", "test-queue")

	assert.Error(t, err)
	assert.Nil(t, subscriber)
	assert.Contains(t, err.Error(), "RabbitMQSubscriber - dial")
}

func TestEvent_Fields(t *testing.T) {
	t.Parallel()

	event := Event{
		ID:      "event-123",
		Type:    "user.created",
		Payload: []byte(`{"user_id": "123"}`),
	}

	assert.Equal(t, "event-123", event.ID)
	assert.Equal(t, "user.created", event.Type)
	assert.Equal(t, []byte(`{"user_id": "123"}`), event.Payload)
}
