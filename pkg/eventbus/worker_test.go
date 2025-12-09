package eventbus_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/evrone/go-clean-template/pkg/eventbus"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errConnectionFailed = errors.New("connection failed")

type mockOutboxRepo struct {
	mu            sync.Mutex
	events        []event.OutboxEvent
	fetchErr      error
	markPubErr    error
	markFailedErr error
	published     []uuid.UUID
	failed        []uuid.UUID
	onPublished   chan uuid.UUID
	onFailed      chan uuid.UUID
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

	for i := range m.events {
		if m.events[i].ID == id {
			m.events = append(m.events[:i], m.events[i+1:]...)

			break
		}
	}

	if m.onPublished != nil {
		select {
		case m.onPublished <- id:
		default:
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

	for i := range m.events {
		if m.events[i].ID == id {
			m.events[i].RetryCount++

			break
		}
	}

	if m.onFailed != nil {
		select {
		case m.onFailed <- id:
		default:
		}
	}

	return nil
}

type mockPublisher struct {
	mu         sync.Mutex
	published  []event.OutboxEvent
	publishErr error
}

func (m *mockPublisher) Publish(_ context.Context, e *event.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.publishErr != nil {
		return m.publishErr
	}

	m.published = append(m.published, *e)

	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{})                {}
func (m *mockLogger) Info(_ string, _ ...interface{})                      {}
func (m *mockLogger) Warn(_ string, _ ...interface{})                      {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{})                {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{})                {}
func (m *mockLogger) WithField(_ string, _ interface{}) logger.Interface   { return m }
func (m *mockLogger) WithFields(_ map[string]interface{}) logger.Interface { return m }
func (m *mockLogger) WithRequestID(_ string) logger.Interface              { return m }
func (m *mockLogger) WithContext(_ context.Context) logger.Interface       { return m }

func TestWorker_ProcessesEvents(t *testing.T) {
	t.Parallel()

	onPublished := make(chan uuid.UUID, 1)
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
		onPublished: onPublished,
	}
	publisher := &mockPublisher{}
	log := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		log,
		eventbus.WithPollInterval(10*time.Millisecond),
		eventbus.WithBatchSize(10),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	select {
	case <-onPublished:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event to be published")
	}

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
	t.Parallel()

	onFailed := make(chan uuid.UUID, 1)
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
		onFailed: onFailed,
	}
	publisher := &mockPublisher{
		publishErr: errConnectionFailed,
	}
	log := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		log,
		eventbus.WithPollInterval(10*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	select {
	case <-onFailed:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event to be marked failed")
	}

	cancel()
	worker.Stop()

	repo.mu.Lock()
	defer repo.mu.Unlock()

	require.GreaterOrEqual(t, len(repo.failed), 1)
	assert.Equal(t, eventID, repo.failed[0])
	assert.Empty(t, repo.published)
}

func TestWorker_SkipsMaxRetries(t *testing.T) {
	t.Parallel()

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
	log := &mockLogger{}

	worker := eventbus.NewWorker(
		repo,
		publisher,
		log,
		eventbus.WithPollInterval(10*time.Millisecond),
		eventbus.WithMaxRetries(5),
	)

	ctx, cancel := context.WithCancel(context.Background())
	worker.Start(ctx)

	// Wait for at least one poll cycle
	time.Sleep(50 * time.Millisecond)

	cancel()
	worker.Stop()

	publisher.mu.Lock()
	defer publisher.mu.Unlock()

	assert.Empty(t, publisher.published)

	repo.mu.Lock()
	defer repo.mu.Unlock()

	assert.Empty(t, repo.published)
}
