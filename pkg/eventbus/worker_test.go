package eventbus_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/evrone/go-clean-template/pkg/eventbus"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOutboxRepo struct {
	mu            sync.Mutex
	events        []event.OutboxEvent
	fetchErr      error
	markPubErr    error
	markFailedErr error
	published     []uuid.UUID
	failed        []uuid.UUID
}

func (m *mockOutboxRepo) Store(_ context.Context, events []event.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, events...)
	return nil
}

func (m *mockOutboxRepo) FetchUnpublished(_ context.Context, limit int) ([]event.OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	if limit > len(m.events) {
		limit = len(m.events)
	}
	return m.events[:limit], nil
}

func (m *mockOutboxRepo) MarkPublished(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.markPubErr != nil {
		return m.markPubErr
	}
	m.published = append(m.published, id)
	// Remove from events to simulate real behavior
	for i, e := range m.events {
		if e.ID == id {
			m.events = append(m.events[:i], m.events[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockOutboxRepo) MarkFailed(_ context.Context, id uuid.UUID, _ error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.markFailedErr != nil {
		return m.markFailedErr
	}
	m.failed = append(m.failed, id)
	// Increment retry count to simulate real behavior
	for i, e := range m.events {
		if e.ID == id {
			m.events[i].RetryCount++
			break
		}
	}
	return nil
}

type mockPublisher struct {
	mu         sync.Mutex
	published  []event.OutboxEvent
	publishErr error
}

func (m *mockPublisher) Publish(_ context.Context, e event.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.publishErr != nil {
		return m.publishErr
	}
	m.published = append(m.published, e)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

func TestWorker_ProcessesEvents(t *testing.T) {
	repo := &mockOutboxRepo{
		events: []event.OutboxEvent{
			{
				ID:            uuid.New(),
				AggregateType: "user",
				AggregateID:   "1",
				EventType:     "user.created",
				Payload:       []byte(`{"user_id":"1"}`),
				CreatedAt:     time.Now(),
			},
		},
	}
	publisher := &mockPublisher{}
	logger := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		logger,
		eventbus.WithPollInterval(50*time.Millisecond),
		eventbus.WithBatchSize(10),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	cancel()
	worker.Stop()

	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	require.Len(t, publisher.published, 1)
	assert.Equal(t, "user.created", publisher.published[0].EventType)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	require.Len(t, repo.published, 1)
}

func TestWorker_HandlesPublishError(t *testing.T) {
	eventID := uuid.New()
	repo := &mockOutboxRepo{
		events: []event.OutboxEvent{
			{
				ID:            eventID,
				AggregateType: "user",
				AggregateID:   "1",
				EventType:     "user.created",
				Payload:       []byte(`{"user_id":"1"}`),
				CreatedAt:     time.Now(),
			},
		},
	}
	publisher := &mockPublisher{
		publishErr: errors.New("connection failed"),
	}
	logger := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		logger,
		eventbus.WithPollInterval(50*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	cancel()
	worker.Stop()

	repo.mu.Lock()
	defer repo.mu.Unlock()
	require.GreaterOrEqual(t, len(repo.failed), 1)
	assert.Equal(t, eventID, repo.failed[0])
	assert.Empty(t, repo.published)
}

func TestWorker_SkipsMaxRetries(t *testing.T) {
	repo := &mockOutboxRepo{
		events: []event.OutboxEvent{
			{
				ID:            uuid.New(),
				AggregateType: "user",
				AggregateID:   "1",
				EventType:     "user.created",
				Payload:       []byte(`{"user_id":"1"}`),
				CreatedAt:     time.Now(),
				RetryCount:    5,
			},
		},
	}
	publisher := &mockPublisher{}
	logger := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		logger,
		eventbus.WithPollInterval(50*time.Millisecond),
		eventbus.WithMaxRetries(5),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	cancel()
	worker.Stop()

	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	assert.Empty(t, publisher.published)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	assert.Empty(t, repo.published)
}
