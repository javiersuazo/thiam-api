package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

type PushTokenRepo struct {
	*postgres.Postgres
}

func NewPushTokenRepo(pg *postgres.Postgres) *PushTokenRepo {
	return &PushTokenRepo{pg}
}

func (r *PushTokenRepo) Store(ctx context.Context, token *notification.PushToken) error {
	now := time.Now().UTC()

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	token.UpdatedAt = now

	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}

	sql := `
		INSERT INTO push_tokens (id, user_id, token, platform, device_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			device_id = EXCLUDED.device_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.Pool.Exec(ctx, sql,
		token.ID, token.UserID, token.Token, token.Platform, token.DeviceID, token.CreatedAt, token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("PushTokenRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *PushTokenRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error) {
	sql, args, err := r.Builder.
		Select("id", "user_id", "token", "platform", "device_id", "created_at", "updated_at").
		From("push_tokens").
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("PushTokenRepo - GetByUserID - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PushTokenRepo - GetByUserID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	tokens := make([]notification.PushToken, 0)

	for rows.Next() {
		var t notification.PushToken

		err = rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.DeviceID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("PushTokenRepo - GetByUserID - rows.Scan: %w", err)
		}

		tokens = append(tokens, t)
	}

	return tokens, nil
}

func (r *PushTokenRepo) Delete(ctx context.Context, token string) error {
	sql, args, err := r.Builder.
		Delete("push_tokens").
		Where("token = ?", token).
		ToSql()
	if err != nil {
		return fmt.Errorf("PushTokenRepo - Delete - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PushTokenRepo - Delete - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *PushTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	sql, args, err := r.Builder.
		Delete("push_tokens").
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("PushTokenRepo - DeleteByUserID - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PushTokenRepo - DeleteByUserID - r.Pool.Exec: %w", err)
	}

	return nil
}
