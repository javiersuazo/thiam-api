# Event-Driven Architecture

This document outlines the event-driven architecture patterns for domain events, notifications (email/SMS), and asynchronous processing.

## Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Aggregate  │───▶│   Outbox    │───▶│  Publisher  │───▶│  Consumer   │
│  (Domain)   │    │   (DB)      │    │  (Worker)   │    │  (Handler)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
      │                  │                  │                  │
      │ Raises           │ Stores           │ Publishes        │ Processes
      │ Events           │ Atomically       │ to RabbitMQ      │ Events
      ▼                  ▼                  ▼                  ▼
  UserCreated       Same TX as         Topic Exchange     SendWelcomeEmail
  OrderPlaced       Aggregate          (fanout)           SendSMS
  PaymentReceived                                         UpdateAnalytics
```

## Transactional Outbox Pattern

The outbox pattern guarantees at-least-once delivery of domain events. Events are stored in the same transaction as the aggregate changes, then published asynchronously.

### Why Outbox?

| Problem | Solution |
|---------|----------|
| Lost events on app crash | Events persisted in DB first |
| Inconsistent state | Same transaction as domain changes |
| Duplicate delivery | Consumer idempotency via event ID |
| Message broker down | Outbox retries automatically |

### Outbox Table Schema

```sql
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    retry_count INT DEFAULT 0,
    last_error TEXT
);

CREATE INDEX idx_outbox_unpublished ON outbox_events (created_at)
    WHERE published_at IS NULL;
```

## Domain Events

### Event Structure

```go
// internal/entity/events/event.go
type Event interface {
    EventID() uuid.UUID
    EventType() string
    AggregateID() uuid.UUID
    AggregateType() string
    OccurredAt() time.Time
    Payload() any
}

type BaseEvent struct {
    ID          uuid.UUID
    Type        string
    AggregateID uuid.UUID
    AggrType    string
    OccurredAt  time.Time
}
```

### Aggregate with Events

```go
// internal/entity/user.go
type User struct {
    ID            uuid.UUID
    Email         string
    EmailVerified bool
    CreatedAt     time.Time

    events []events.Event // uncommitted events
}

func NewUser(email string) (*User, error) {
    if !isValidEmail(email) {
        return nil, ErrInvalidEmail
    }

    user := &User{
        ID:        uuid.New(),
        Email:     email,
        CreatedAt: time.Now(),
    }

    user.raise(events.UserCreated{
        UserID: user.ID,
        Email:  user.Email,
    })

    return user, nil
}

func (u *User) raise(event events.Event) {
    u.events = append(u.events, event)
}

func (u *User) Events() []events.Event {
    return u.events
}

func (u *User) ClearEvents() {
    u.events = nil
}
```

### Event Examples

```go
// internal/entity/events/user_events.go
type UserCreated struct {
    BaseEvent
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
}

func (e UserCreated) EventType() string     { return "user.created" }
func (e UserCreated) AggregateType() string { return "user" }

type UserEmailVerified struct {
    BaseEvent
    UserID uuid.UUID `json:"user_id"`
}

func (e UserEmailVerified) EventType() string     { return "user.email_verified" }
func (e UserEmailVerified) AggregateType() string { return "user" }
```

## Repository Pattern with Events

```go
// internal/repo/contracts.go
type UserRepository interface {
    Save(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

// internal/repo/user_pg.go
func (r *UserRepo) Save(ctx context.Context, user *entity.User) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("UserRepo.Save begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    // Save aggregate
    _, err = tx.Exec(ctx, upsertUserSQL, user.ID, user.Email, user.CreatedAt)
    if err != nil {
        return fmt.Errorf("UserRepo.Save upsert: %w", err)
    }

    // Save events to outbox (same transaction)
    for _, event := range user.Events() {
        payload, _ := json.Marshal(event.Payload())
        _, err = tx.Exec(ctx, insertOutboxSQL,
            event.EventID(),
            event.AggregateType(),
            event.AggregateID(),
            event.EventType(),
            payload,
        )
        if err != nil {
            return fmt.Errorf("UserRepo.Save outbox: %w", err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("UserRepo.Save commit: %w", err)
    }

    user.ClearEvents()
    return nil
}
```

## Outbox Worker

The outbox worker polls for unpublished events and publishes them to RabbitMQ:

```go
// internal/infrastructure/outbox/worker.go
type Worker struct {
    pool      *pgxpool.Pool
    publisher EventPublisher
    interval  time.Duration
    batchSize int
}

func (w *Worker) Run(ctx context.Context) error {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := w.processOutbox(ctx); err != nil {
                log.Error().Err(err).Msg("outbox processing failed")
            }
        }
    }
}

func (w *Worker) processOutbox(ctx context.Context) error {
    events, err := w.fetchUnpublished(ctx, w.batchSize)
    if err != nil {
        return err
    }

    for _, event := range events {
        if err := w.publisher.Publish(ctx, event); err != nil {
            w.recordError(ctx, event.ID, err)
            continue
        }
        w.markPublished(ctx, event.ID)
    }

    return nil
}
```

## Event Bus (RabbitMQ)

### Exchange Configuration

```go
// Topic exchange for routing events to interested consumers
exchange := "domain.events"
exchangeType := "topic"

// Routing keys follow pattern: aggregate.event
// Examples:
// - user.created
// - user.email_verified
// - order.placed
// - payment.received
```

### Publisher Interface

```go
// internal/infrastructure/eventbus/contracts.go
type EventPublisher interface {
    Publish(ctx context.Context, event OutboxEvent) error
}

// internal/infrastructure/eventbus/rabbitmq.go
type RabbitMQPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, event OutboxEvent) error {
    body, _ := json.Marshal(event)

    return p.channel.PublishWithContext(ctx,
        "domain.events",           // exchange
        event.EventType,           // routing key (e.g., "user.created")
        false,                     // mandatory
        false,                     // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: amqp.Persistent,
            MessageId:    event.ID.String(),
            Timestamp:    event.CreatedAt,
            Body:         body,
        },
    )
}
```

## Event Consumers

### Consumer Interface

```go
// internal/infrastructure/eventbus/contracts.go
type EventHandler interface {
    Handle(ctx context.Context, event EventMessage) error
    EventTypes() []string // which events this handler processes
}
```

### Notification Service

```go
// internal/usecase/notification/service.go
type Service struct {
    emailSender EmailSender
    smsSender   SMSSender
    templates   TemplateRenderer
}

func (s *Service) Handle(ctx context.Context, event EventMessage) error {
    switch event.Type {
    case "user.created":
        return s.handleUserCreated(ctx, event)
    case "order.placed":
        return s.handleOrderPlaced(ctx, event)
    default:
        return nil // ignore unknown events
    }
}

func (s *Service) handleUserCreated(ctx context.Context, event EventMessage) error {
    var payload UserCreatedPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    body, err := s.templates.Render("welcome_email", payload)
    if err != nil {
        return fmt.Errorf("render template: %w", err)
    }

    return s.emailSender.Send(ctx, Email{
        To:      payload.Email,
        Subject: "Welcome!",
        Body:    body,
    })
}

func (s *Service) EventTypes() []string {
    return []string{"user.created", "order.placed", "payment.received"}
}
```

## Notification Adapters

### Email Interface

```go
// internal/usecase/notification/email.go
type EmailSender interface {
    Send(ctx context.Context, email Email) error
}

type Email struct {
    To      string
    Subject string
    Body    string
    HTML    bool
}
```

### SMS Interface

```go
// internal/usecase/notification/sms.go
type SMSSender interface {
    Send(ctx context.Context, sms SMS) error
}

type SMS struct {
    To   string
    Body string
}
```

### Adapter Implementations

```go
// internal/infrastructure/notification/sendgrid.go
type SendGridEmailSender struct {
    client *sendgrid.Client
    from   string
}

func (s *SendGridEmailSender) Send(ctx context.Context, email Email) error {
    // SendGrid implementation
}

// internal/infrastructure/notification/twilio.go
type TwilioSMSSender struct {
    client *twilio.Client
    from   string
}

func (t *TwilioSMSSender) Send(ctx context.Context, sms SMS) error {
    // Twilio implementation
}

// internal/infrastructure/notification/log.go (for development)
type LogEmailSender struct {
    log zerolog.Logger
}

func (l *LogEmailSender) Send(ctx context.Context, email Email) error {
    l.log.Info().
        Str("to", email.To).
        Str("subject", email.Subject).
        Msg("Email sent (logged)")
    return nil
}
```

## Idempotency

Consumers must handle duplicate events:

```go
// internal/infrastructure/eventbus/consumer.go
type IdempotentConsumer struct {
    handler    EventHandler
    processed  ProcessedEventStore
}

func (c *IdempotentConsumer) Handle(ctx context.Context, event EventMessage) error {
    // Check if already processed
    if c.processed.Exists(ctx, event.ID) {
        return nil // skip duplicate
    }

    // Process event
    if err := c.handler.Handle(ctx, event); err != nil {
        return err
    }

    // Mark as processed
    return c.processed.Store(ctx, event.ID)
}
```

## Configuration

```bash
# Event Bus
EVENT_BUS_EXCHANGE=domain.events
EVENT_BUS_URL=amqp://guest:guest@localhost:5672/

# Outbox Worker
OUTBOX_POLL_INTERVAL=1s
OUTBOX_BATCH_SIZE=100
OUTBOX_MAX_RETRIES=5

# Email (SendGrid)
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM=noreply@example.com

# SMS (Twilio)
TWILIO_ACCOUNT_SID=ACxxx
TWILIO_AUTH_TOKEN=xxx
TWILIO_FROM=+1234567890

# Development (use log adapters)
EMAIL_ADAPTER=log
SMS_ADAPTER=log
```

## Testing

### Unit Testing Events

```go
func TestUser_Creation_RaisesEvent(t *testing.T) {
    user, err := entity.NewUser("test@example.com")
    require.NoError(t, err)

    events := user.Events()
    require.Len(t, events, 1)

    created, ok := events[0].(events.UserCreated)
    require.True(t, ok)
    assert.Equal(t, user.ID, created.UserID)
    assert.Equal(t, "test@example.com", created.Email)
}
```

### Integration Testing Outbox

```go
func TestOutboxWorker_PublishesEvents(t *testing.T) {
    // Setup test database and mock publisher
    db := setupTestDB(t)
    publisher := &mockPublisher{}
    worker := outbox.NewWorker(db, publisher, 100*time.Millisecond, 10)

    // Insert test event
    insertOutboxEvent(t, db, events.UserCreated{...})

    // Run worker
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    go worker.Run(ctx)

    // Assert published
    assert.Eventually(t, func() bool {
        return len(publisher.published) == 1
    }, time.Second, 100*time.Millisecond)
}
```

## Implementation Phases

### Phase 1: Foundation
- [ ] Create `outbox_events` migration
- [ ] Implement domain event interfaces
- [ ] Add events to User aggregate
- [ ] Update UserRepository to save events

### Phase 2: Outbox Worker
- [ ] Implement outbox worker
- [ ] Configure RabbitMQ topic exchange
- [ ] Implement RabbitMQ publisher

### Phase 3: Notification Service
- [ ] Implement notification service
- [ ] Create email sender interface + adapters
- [ ] Create SMS sender interface + adapters
- [ ] Implement welcome email handler

### Phase 4: Production Readiness
- [ ] Add idempotency layer
- [ ] Implement dead-letter queue
- [ ] Add metrics and monitoring
- [ ] Configure retry policies
