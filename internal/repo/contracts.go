// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=mocks/repo_mock.go -package=mocks

type (
	// OutboxRepo handles outbox event persistence.
	OutboxRepo interface {
		Store(ctx context.Context, events []event.OutboxEvent) error
		FetchUnpublished(ctx context.Context, limit int) ([]event.OutboxEvent, error)
		MarkPublished(ctx context.Context, id uuid.UUID) error
		MarkFailed(ctx context.Context, id uuid.UUID, err error) error
	}

	// NotificationRepo handles in-app notification persistence.
	NotificationRepo interface {
		Store(ctx context.Context, n *notification.InAppNotification) error
		GetByID(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error)
		GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]notification.InAppNotification, error)
		MarkAsRead(ctx context.Context, id uuid.UUID) error
		MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
		GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	}

	// NotificationPreferencesRepo handles user notification preferences.
	NotificationPreferencesRepo interface {
		Get(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error)
		Upsert(ctx context.Context, prefs *notification.UserPreferences) error
	}

	// PushTokenRepo handles push notification tokens.
	PushTokenRepo interface {
		Store(ctx context.Context, token *notification.PushToken) error
		GetByUserID(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error)
		Delete(ctx context.Context, token string) error
		DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	}

	// DeliveryLogRepo handles notification delivery logs.
	DeliveryLogRepo interface {
		Store(ctx context.Context, log *notification.DeliveryLog) error
		GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]notification.DeliveryLog, error)
	}
)
