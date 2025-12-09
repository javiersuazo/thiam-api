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

type mockPreferencesRepo struct {
	getFunc    func(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error)
	upsertFunc func(ctx context.Context, prefs *notification.UserPreferences) error
}

func (m *mockPreferencesRepo) Get(ctx context.Context, userID uuid.UUID) (*notification.UserPreferences, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}

	return nil, nil
}

func (m *mockPreferencesRepo) Upsert(ctx context.Context, prefs *notification.UserPreferences) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, prefs)
	}

	return nil
}

func TestPreferencesUseCase_Get(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockPreferencesRepo
		userID  uuid.UUID
		want    *notification.UserPreferences
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{
						UserID:       userID,
						EmailEnabled: true,
						PushEnabled:  true,
						InAppEnabled: true,
					}, nil
				},
			},
			userID: userID,
			want: &notification.UserPreferences{
				UserID:       userID,
				EmailEnabled: true,
				PushEnabled:  true,
				InAppEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
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

			uc := notificationuc.NewPreferencesUseCase(tt.repo)
			got, err := uc.Get(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPreferencesUseCase_Update(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockPreferencesRepo
		input   *notification.UserPreferences
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockPreferencesRepo{},
			input: &notification.UserPreferences{
				UserID:       userID,
				EmailEnabled: false,
				PushEnabled:  true,
			},
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockPreferencesRepo{
				upsertFunc: func(_ context.Context, _ *notification.UserPreferences) error {
					return errRepo
				},
			},
			input: &notification.UserPreferences{
				UserID: userID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewPreferencesUseCase(tt.repo)
			err := uc.Update(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
