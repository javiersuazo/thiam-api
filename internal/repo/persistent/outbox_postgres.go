package persistent

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

type OutboxRepo struct {
	*postgres.Postgres
}

func NewOutboxRepo(pg *postgres.Postgres) *OutboxRepo {
	return &OutboxRepo{pg}
}

func (r *OutboxRepo) Store(ctx context.Context, events []event.OutboxEvent) error {
	if len(events) == 0 {
		return nil
	}

	query := r.Builder.
		Insert("outbox_events").
		Columns("id", "aggregate_type", "aggregate_id", "event_type", "payload", "created_at")

	for i := range events {
		query = query.Values(events[i].ID, events[i].AggregateType, events[i].AggregateID, events[i].EventType, events[i].Payload, events[i].CreatedAt)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("OutboxRepo.Store - build query: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxRepo.Store - exec: %w", err)
	}

	return nil
}

func (r *OutboxRepo) FetchUnpublished(ctx context.Context, limit int) ([]event.OutboxEvent, error) {
	if limit < 0 {
		limit = 0
	}

	sql, args, err := r.Builder.
		Select("id", "aggregate_type", "aggregate_id", "event_type", "payload", "created_at", "retry_count", "last_error").
		From("outbox_events").
		Where("published_at IS NULL").
		OrderBy("created_at ASC").
		Limit(uint64(limit)). //nolint:gosec // limit is checked above
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("OutboxRepo.FetchUnpublished - build query: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("OutboxRepo.FetchUnpublished - query: %w", err)
	}
	defer rows.Close()

	events := make([]event.OutboxEvent, 0, limit)

	for rows.Next() {
		var e event.OutboxEvent

		err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.CreatedAt, &e.RetryCount, &e.LastError)
		if err != nil {
			return nil, fmt.Errorf("OutboxRepo.FetchUnpublished - scan: %w", err)
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("OutboxRepo.FetchUnpublished - rows: %w", err)
	}

	return events, nil
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.Builder.
		Update("outbox_events").
		Set("published_at", time.Now().UTC()).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxRepo.MarkPublished - build query: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxRepo.MarkPublished - exec: %w", err)
	}

	return nil
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, publishErr error) error {
	errMsg := publishErr.Error()

	sql, args, err := r.Builder.
		Update("outbox_events").
		Set("retry_count", sq.Expr("retry_count + 1")).
		Set("last_error", errMsg).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("OutboxRepo.MarkFailed - build query: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("OutboxRepo.MarkFailed - exec: %w", err)
	}

	return nil
}
