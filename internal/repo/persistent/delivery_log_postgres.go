package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

type DeliveryLogRepo struct {
	*postgres.Postgres
}

func NewDeliveryLogRepo(pg *postgres.Postgres) *DeliveryLogRepo {
	return &DeliveryLogRepo{pg}
}

func (r *DeliveryLogRepo) Store(ctx context.Context, log *notification.DeliveryLog) error {
	now := time.Now().UTC()

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}

	sql, args, err := r.Builder.
		Insert("notification_delivery_logs").
		Columns("id", "notification_id", "user_id", "channel", "status", "provider", "provider_message_id", "error_message", "attempts", "created_at", "delivered_at").
		Values(log.ID, log.NotificationID, log.UserID, log.Channel, log.Status, log.Provider, log.ProviderMsgID, log.ErrorMessage, log.Attempts, log.CreatedAt, log.DeliveredAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("DeliveryLogRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeliveryLogRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *DeliveryLogRepo) GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]notification.DeliveryLog, error) {
	sql, args, err := r.Builder.
		Select("id", "notification_id", "user_id", "channel", "status", "provider", "provider_message_id", "error_message", "attempts", "created_at", "delivered_at").
		From("notification_delivery_logs").
		Where("notification_id = ?", notificationID).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("DeliveryLogRepo - GetByNotificationID - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("DeliveryLogRepo - GetByNotificationID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	logs := make([]notification.DeliveryLog, 0)

	for rows.Next() {
		var l notification.DeliveryLog

		err = rows.Scan(&l.ID, &l.NotificationID, &l.UserID, &l.Channel, &l.Status, &l.Provider, &l.ProviderMsgID, &l.ErrorMessage, &l.Attempts, &l.CreatedAt, &l.DeliveredAt)
		if err != nil {
			return nil, fmt.Errorf("DeliveryLogRepo - GetByNotificationID - rows.Scan: %w", err)
		}

		logs = append(logs, l)
	}

	return logs, nil
}
