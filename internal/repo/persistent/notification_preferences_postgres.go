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

var errPreferencesNotFound = errors.New("notification preferences not found")

type NotificationPreferencesRepo struct {
	*postgres.Postgres
}

func NewNotificationPreferencesRepo(pg *postgres.Postgres) *NotificationPreferencesRepo {
	return &NotificationPreferencesRepo{pg}
}

func (r *NotificationPreferencesRepo) Get(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error) {
	sql, args, err := r.Builder.
		Select("user_id", "email_enabled", "sms_enabled", "push_enabled", "in_app_enabled", "quiet_start", "quiet_end", "updated_at").
		From("notification_preferences").
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("NotificationPreferencesRepo - Get - r.Builder: %w", err)
	}

	var p notification.UserPreferences

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&p.UserID, &p.EmailEnabled, &p.SMSEnabled, &p.PushEnabled, &p.InAppEnabled,
		&p.QuietStart, &p.QuietEnd, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errPreferencesNotFound
		}

		return nil, fmt.Errorf("NotificationPreferencesRepo - Get - r.Pool.QueryRow: %w", err)
	}

	return &p, nil
}

func (r *NotificationPreferencesRepo) Upsert(ctx context.Context, prefs *notification.UserPreferences) error {
	now := time.Now().UTC()
	prefs.UpdatedAt = now

	sql := `
		INSERT INTO notification_preferences (user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled, quiet_start, quiet_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			sms_enabled = EXCLUDED.sms_enabled,
			push_enabled = EXCLUDED.push_enabled,
			in_app_enabled = EXCLUDED.in_app_enabled,
			quiet_start = EXCLUDED.quiet_start,
			quiet_end = EXCLUDED.quiet_end,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.Pool.Exec(ctx, sql,
		prefs.UserID, prefs.EmailEnabled, prefs.SMSEnabled, prefs.PushEnabled, prefs.InAppEnabled,
		prefs.QuietStart, prefs.QuietEnd, now, prefs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("NotificationPreferencesRepo - Upsert - r.Pool.Exec: %w", err)
	}

	return nil
}
