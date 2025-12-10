// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/internal/entity/notification"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// AuthUseCase handles authentication operations.
	AuthUseCase interface {
		Register(ctx context.Context, input *authuc.RegisterInput) (*authuc.RegisterOutput, error)
		Login(ctx context.Context, input *authuc.LoginInput) (*authuc.LoginOutput, error)
		Logout(ctx context.Context, refreshToken string) error
		LogoutAll(ctx context.Context, userID uuid.UUID) error
		RefreshToken(ctx context.Context, input *authuc.RefreshTokenInput) (*authuc.RefreshTokenOutput, error)
		GetCurrentUser(ctx context.Context, userID uuid.UUID) (*auth.User, error)
		ValidateAccessToken(tokenString string) (*pkgauth.Claims, error)
	}

	// InAppNotificationUseCase handles in-app notification operations.
	InAppNotificationUseCase interface {
		Create(ctx context.Context, n *notification.InAppNotification) error
		GetByID(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error)
		GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]notification.InAppNotification, error)
		MarkAsRead(ctx context.Context, id uuid.UUID) error
		MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
		GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	}

	// NotificationPreferences handles user notification preferences.
	NotificationPreferences interface {
		Get(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error)
		Update(ctx context.Context, prefs *notification.UserPreferences) error
	}

	// PushToken handles push notification token management.
	PushToken interface {
		Register(ctx context.Context, token *notification.PushToken) error
		GetByUserID(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error)
		Unregister(ctx context.Context, token string) error
		UnregisterAll(ctx context.Context, userID uuid.UUID) error
	}
)
