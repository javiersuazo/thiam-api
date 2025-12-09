package notification

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/google/uuid"
)

type PreferencesUseCase struct {
	repo repo.NotificationPreferencesRepo
}

func NewPreferencesUseCase(r repo.NotificationPreferencesRepo) *PreferencesUseCase {
	return &PreferencesUseCase{
		repo: r,
	}
}

func (uc *PreferencesUseCase) Get(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error) {
	prefs, err := uc.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("PreferencesUseCase - Get - uc.repo.Get: %w", err)
	}

	return prefs, nil
}

func (uc *PreferencesUseCase) Update(ctx context.Context, prefs *notification.UserPreferences) error {
	if err := uc.repo.Upsert(ctx, prefs); err != nil {
		return fmt.Errorf("PreferencesUseCase - Update - uc.repo.Upsert: %w", err)
	}

	return nil
}
