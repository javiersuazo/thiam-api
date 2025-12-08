# Event-Driven Patterns

This document describes event-driven patterns that **complement** the primary API. Events handle asynchronous side-effects after API operations complete.

## When to Use Events vs Direct API Calls

| Use Case | Pattern | Example |
|----------|---------|---------|
| User requests data | **API (sync)** | `GET /v1/users/:id` |
| User creates resource | **API (sync)** | `POST /v1/users` |
| Send welcome email after signup | **Event (async)** | `UserCreated` → Email service |
| Update analytics after purchase | **Event (async)** | `OrderPlaced` → Analytics service |
| Notify external systems | **Event (async)** | `PaymentReceived` → Webhook |

**Rule**: The API operation completes synchronously. Side-effects happen asynchronously via events.

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   API Request   │────▶│    Use Case     │────▶│    Response     │
│   POST /users   │     │  Create User    │     │   201 Created   │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 │ (same transaction)
                                 ▼
                        ┌─────────────────┐
                        │  Outbox Table   │
                        │  UserCreated    │
                        └────────┬────────┘
                                 │
                                 │ (async, later)
                                 ▼
                        ┌─────────────────┐
                        │  Outbox Worker  │
                        └────────┬────────┘
                                 │
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
             ┌──────────┐ ┌──────────┐ ┌──────────┐
             │  Email   │ │   SMS    │ │Analytics │
             │ Service  │ │ Service  │ │ Service  │
             └──────────┘ └──────────┘ └──────────┘
```

## Transactional Outbox Pattern

### Why Outbox?

The outbox pattern guarantees event delivery without distributed transactions:

1. API creates/updates entity in database
2. Event is written to outbox table **in the same transaction**
3. Worker polls outbox and publishes to message queue
4. Consumers process events asynchronously

If the API fails, no event is written. If publishing fails, the worker retries.

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

### Event Interface

```go
type Event interface {
    EventID() uuid.UUID
    EventType() string
    AggregateID() uuid.UUID
    AggregateType() string
    OccurredAt() time.Time
}
```

### Event Examples

```go
type UserCreated struct {
    ID          uuid.UUID `json:"id"`
    UserID      uuid.UUID `json:"user_id"`
    Email       string    `json:"email"`
    OccurredAt  time.Time `json:"occurred_at"`
}

func (e UserCreated) EventType() string     { return "user.created" }
func (e UserCreated) AggregateType() string { return "user" }

type OrderPlaced struct {
    ID         uuid.UUID `json:"id"`
    OrderID    uuid.UUID `json:"order_id"`
    UserID     uuid.UUID `json:"user_id"`
    Total      int64     `json:"total"`
    OccurredAt time.Time `json:"occurred_at"`
}

func (e OrderPlaced) EventType() string     { return "order.placed" }
func (e OrderPlaced) AggregateType() string { return "order" }
```

## Aggregate Pattern with Events

Aggregates collect events that are persisted atomically with the aggregate state:

```go
type User struct {
    ID        uuid.UUID
    Email     string
    CreatedAt time.Time

    events []Event
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

    user.raise(UserCreated{
        ID:         uuid.New(),
        UserID:     user.ID,
        Email:      user.Email,
        OccurredAt: time.Now(),
    })

    return user, nil
}

func (u *User) raise(event Event) {
    u.events = append(u.events, event)
}

func (u *User) Events() []Event {
    return u.events
}

func (u *User) ClearEvents() {
    u.events = nil
}
```

## Repository Saves Events Atomically

```go
func (r *UserRepo) Save(ctx context.Context, user *entity.User) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("UserRepo.Save - begin: %w", err)
    }
    defer tx.Rollback(ctx)

    _, err = tx.Exec(ctx, insertUserSQL, user.ID, user.Email, user.CreatedAt)
    if err != nil {
        return fmt.Errorf("UserRepo.Save - insert user: %w", err)
    }

    for _, event := range user.Events() {
        payload, _ := json.Marshal(event)
        _, err = tx.Exec(ctx, insertOutboxSQL,
            event.EventID(),
            event.AggregateType(),
            event.AggregateID(),
            event.EventType(),
            payload,
        )
        if err != nil {
            return fmt.Errorf("UserRepo.Save - insert outbox: %w", err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("UserRepo.Save - commit: %w", err)
    }

    user.ClearEvents()
    return nil
}
```

## Notification Services

### Email Interface

```go
type EmailSender interface {
    Send(ctx context.Context, email Email) error
}

type Email struct {
    To      string
    Subject string
    Body    string
}
```

### SMS Interface

```go
type SMSSender interface {
    Send(ctx context.Context, sms SMS) error
}

type SMS struct {
    To   string
    Body string
}
```

### Adapter Pattern

Production adapters implement the interfaces:

```go
type SendGridEmailSender struct {
    client *sendgrid.Client
    from   string
}

func (s *SendGridEmailSender) Send(ctx context.Context, email Email) error {
    // SendGrid API call
}

type TwilioSMSSender struct {
    client *twilio.Client
    from   string
}

func (t *TwilioSMSSender) Send(ctx context.Context, sms SMS) error {
    // Twilio API call
}
```

Development adapter for local testing:

```go
type LogEmailSender struct {
    log zerolog.Logger
}

func (l *LogEmailSender) Send(ctx context.Context, email Email) error {
    l.log.Info().
        Str("to", email.To).
        Str("subject", email.Subject).
        Msg("email sent (logged)")
    return nil
}
```

## Event Consumer

```go
type NotificationHandler struct {
    emailSender EmailSender
    smsSender   SMSSender
}

func (h *NotificationHandler) Handle(ctx context.Context, event EventMessage) error {
    switch event.Type {
    case "user.created":
        return h.handleUserCreated(ctx, event)
    case "order.placed":
        return h.handleOrderPlaced(ctx, event)
    default:
        return nil
    }
}

func (h *NotificationHandler) handleUserCreated(ctx context.Context, event EventMessage) error {
    var payload UserCreatedPayload
    if err := json.Unmarshal(event.Payload, &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    return h.emailSender.Send(ctx, Email{
        To:      payload.Email,
        Subject: "Welcome!",
        Body:    "Thanks for signing up.",
    })
}
```

## Idempotency

Consumers must handle duplicate events:

```go
type IdempotentHandler struct {
    handler   EventHandler
    processed ProcessedEventStore
}

func (h *IdempotentHandler) Handle(ctx context.Context, event EventMessage) error {
    if h.processed.Exists(ctx, event.ID) {
        return nil
    }

    if err := h.handler.Handle(ctx, event); err != nil {
        return err
    }

    return h.processed.Store(ctx, event.ID)
}
```

## Configuration

```bash
# Outbox Worker
OUTBOX_POLL_INTERVAL=1s
OUTBOX_BATCH_SIZE=100
OUTBOX_MAX_RETRIES=5

# Email
EMAIL_ADAPTER=sendgrid  # or "log" for development
SENDGRID_API_KEY=SG.xxx
SENDGRID_FROM=noreply@example.com

# SMS
SMS_ADAPTER=twilio  # or "log" for development
TWILIO_ACCOUNT_SID=ACxxx
TWILIO_AUTH_TOKEN=xxx
TWILIO_FROM=+1234567890
```

## Implementation Phases

### Phase 1: Foundation
- Create outbox_events migration
- Define domain event interface
- Update aggregates to raise events
- Update repositories to save events atomically

### Phase 2: Outbox Worker
- Implement polling worker
- Configure RabbitMQ topic exchange
- Implement publisher

### Phase 3: Notifications
- Define email/SMS interfaces
- Implement SendGrid adapter
- Implement Twilio adapter
- Implement log adapter (dev)
- Create notification handler

### Phase 4: Production Hardening
- Add idempotency layer
- Implement dead-letter queue
- Add metrics/monitoring
- Configure retry policies
