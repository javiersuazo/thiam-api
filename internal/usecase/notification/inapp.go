package notification

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

type InAppUseCase struct {
	repo repo.NotificationRepo
}

func NewInAppUseCase(r repo.NotificationRepo) *InAppUseCase {
	return &InAppUseCase{
		repo: r,
	}
}

func (uc *InAppUseCase) Create(ctx context.Context, n *notification.InAppNotification) error {
	if err := uc.repo.Store(ctx, n); err != nil {
		return fmt.Errorf("InAppUseCase - Create - uc.repo.Store: %w", err)
	}

	return nil
}

func (uc *InAppUseCase) GetByID(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error) {
	n, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("InAppUseCase - GetByID - uc.repo.GetByID: %w", err)
	}

	return n, nil
}

func (uc *InAppUseCase) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]notification.InAppNotification, error) {
	notifications, err := uc.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("InAppUseCase - GetByUserID - uc.repo.GetByUserID: %w", err)
	}

	return notifications, nil
}

func (uc *InAppUseCase) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if err := uc.repo.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("InAppUseCase - MarkAsRead - uc.repo.MarkAsRead: %w", err)
	}

	return nil
}

func (uc *InAppUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("InAppUseCase - MarkAllAsRead - uc.repo.MarkAllAsRead: %w", err)
	}

	return nil
}

func (uc *InAppUseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := uc.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("InAppUseCase - GetUnreadCount - uc.repo.GetUnreadCount: %w", err)
	}

	return count, nil
}
