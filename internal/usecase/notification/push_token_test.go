package notification_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	notificationuc "github.com/evrone/go-clean-template/internal/usecase/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPushTokenRepo struct {
	storeFunc          func(ctx context.Context, token *notification.PushToken) error
	getByUserIDFunc    func(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error)
	deleteFunc         func(ctx context.Context, token string) error
	deleteByUserIDFunc func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockPushTokenRepo) Store(ctx context.Context, token *notification.PushToken) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, token)
	}

	return nil
}

func (m *mockPushTokenRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]notification.PushToken, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}

	return nil, nil
}

func (m *mockPushTokenRepo) Delete(ctx context.Context, token string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, token)
	}

	return nil
}

func (m *mockPushTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.deleteByUserIDFunc != nil {
		return m.deleteByUserIDFunc(ctx, userID)
	}

	return nil
}

func TestPushTokenUseCase_Register(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockPushTokenRepo
		input   *notification.PushToken
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockPushTokenRepo{},
			input: &notification.PushToken{
				UserID:   userID,
				Token:    "fcm-token-123",
				Platform: "android",
				DeviceID: "device-123",
			},
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPushTokenRepo{
				storeFunc: func(_ context.Context, _ *notification.PushToken) error {
					return errRepo
				},
			},
			input: &notification.PushToken{
				UserID: userID,
				Token:  "fcm-token-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewPushTokenUseCase(tt.repo)
			err := uc.Register(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

//nolint:funlen // table-driven tests are verbose
func TestPushTokenUseCase_GetByUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockPushTokenRepo
		userID  uuid.UUID
		want    []notification.PushToken
		wantErr bool
	}{
		{
			name: "success with tokens",
			repo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{
						{UserID: userID, Token: "token-1", Platform: "ios"},
						{UserID: userID, Token: "token-2", Platform: "android"},
					}, nil
				},
			},
			userID: userID,
			want: []notification.PushToken{
				{UserID: userID, Token: "token-1", Platform: "ios"},
				{UserID: userID, Token: "token-2", Platform: "android"},
			},
			wantErr: false,
		},
		{
			name: "empty result",
			repo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{}, nil
				},
			},
			userID:  userID,
			want:    []notification.PushToken{},
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return nil, errRepo
				},
			},
			userID:  userID,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewPushTokenUseCase(tt.repo)
			got, err := uc.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPushTokenUseCase_Unregister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    *mockPushTokenRepo
		token   string
		wantErr bool
	}{
		{
			name:    "success",
			repo:    &mockPushTokenRepo{},
			token:   "fcm-token-123",
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPushTokenRepo{
				deleteFunc: func(_ context.Context, _ string) error {
					return errRepo
				},
			},
			token:   "fcm-token-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewPushTokenUseCase(tt.repo)
			err := uc.Unregister(context.Background(), tt.token)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPushTokenUseCase_UnregisterAll(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockPushTokenRepo
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "success",
			repo:    &mockPushTokenRepo{},
			userID:  userID,
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPushTokenRepo{
				deleteByUserIDFunc: func(_ context.Context, _ uuid.UUID) error {
					return errRepo
				},
			},
			userID:  userID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewPushTokenUseCase(tt.repo)
			err := uc.UnregisterAll(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
