package notification

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

type PushTokenUseCase struct {
	repo repo.PushTokenRepo
}

func NewPushTokenUseCase(r repo.PushTokenRepo) *PushTokenUseCase {
	return &PushTokenUseCase{
		repo: r,
	}
}

func (uc *PushTokenUseCase) Register(ctx context.Context, token *notification.PushToken) error {
	if err := uc.repo.Store(ctx, token); err != nil {
		return fmt.Errorf("PushTokenUseCase - Register - uc.repo.Store: %w", err)
	}

	return nil
}

func (uc *PushTokenUseCase) GetByUserID(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error) {
	tokens, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("PushTokenUseCase - GetByUserID - uc.repo.GetByUserID: %w", err)
	}

	return tokens, nil
}

func (uc *PushTokenUseCase) Unregister(ctx context.Context, token string) error {
	if err := uc.repo.Delete(ctx, token); err != nil {
		return fmt.Errorf("PushTokenUseCase - Unregister - uc.repo.Delete: %w", err)
	}

	return nil
}

func (uc *PushTokenUseCase) UnregisterAll(ctx context.Context, userID uuid.UUID) error {
	if err := uc.repo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("PushTokenUseCase - UnregisterAll - uc.repo.DeleteByUserID: %w", err)
	}

	return nil
}
