package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errNotificationNotFound = errors.New("notification not found")

type NotificationRepo struct {
	*postgres.Postgres
}

func NewNotificationRepo(pg *postgres.Postgres) *NotificationRepo {
	return &NotificationRepo{pg}
}

func (r *NotificationRepo) Store(ctx context.Context, n *notification.InAppNotification) error {
	now := time.Now().UTC()

	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	n.CreatedAt = now

	sql, args, err := r.Builder.
		Insert("notifications").
		Columns("id", "user_id", "type", "title", "body", "data", "action_url", "image_url", "read", "read_at", "created_at").
		Values(n.ID, n.UserID, n.Type, n.Title, n.Body, n.Data, n.ActionURL, n.ImageURL, n.Read, n.ReadAt, n.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error) {
	sql, args, err := r.Builder.
		Select("id", "user_id", "type", "title", "body", "data", "action_url", "image_url", "read", "read_at", "created_at").
		From("notifications").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("NotificationRepo - GetByID - r.Builder: %w", err)
	}

	var n notification.InAppNotification

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.ActionURL, &n.ImageURL, &n.Read, &n.ReadAt, &n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errNotificationNotFound
		}

		return nil, fmt.Errorf("NotificationRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return &n, nil
}

func (r *NotificationRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]notification.InAppNotification, error) {
	sql, args, err := r.Builder.
		Select("id", "user_id", "type", "title", "body", "data", "action_url", "image_url", "read", "read_at", "created_at").
		From("notifications").
		Where("user_id = ?", userID).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("NotificationRepo - GetByUserID - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("NotificationRepo - GetByUserID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	notifications := make([]notification.InAppNotification, 0)

	for rows.Next() {
		var n notification.InAppNotification

		err = rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.ActionURL, &n.ImageURL, &n.Read, &n.ReadAt, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("NotificationRepo - GetByUserID - rows.Scan: %w", err)
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	sql, args, err := r.Builder.
		Update("notifications").
		Set("read", true).
		Set("read_at", now).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAsRead - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAsRead - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	sql, args, err := r.Builder.
		Update("notifications").
		Set("read", true).
		Set("read_at", now).
		Where("user_id = ? AND read = false", userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAllAsRead - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("NotificationRepo - MarkAllAsRead - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	sql, args, err := r.Builder.
		Select("COUNT(*)").
		From("notifications").
		Where("user_id = ? AND read = false", userID).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("NotificationRepo - GetUnreadCount - r.Builder: %w", err)
	}

	var count int

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("NotificationRepo - GetUnreadCount - r.Pool.QueryRow: %w", err)
	}

	return count, nil
}
