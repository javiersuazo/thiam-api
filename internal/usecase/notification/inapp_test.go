package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	notificationuc "github.com/evrone/go-clean-template/internal/usecase/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRepo = errors.New("repository error")

type mockNotificationRepo struct {
	storeFunc          func(ctx context.Context, n *notification.InAppNotification) error
	getByIDFunc        func(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error)
	getByUserIDFunc    func(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]notification.InAppNotification, error)
	markAsReadFunc     func(ctx context.Context, id uuid.UUID) error
	markAllAsReadFunc  func(ctx context.Context, userID uuid.UUID) error
	getUnreadCountFunc func(ctx context.Context, userID uuid.UUID) (int, error)
}

func (m *mockNotificationRepo) Store(ctx context.Context, n *notification.InAppNotification) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, n)
	}

	return nil
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*notification.InAppNotification, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}

	return nil, nil
}

func (m *mockNotificationRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset uint64) ([]notification.InAppNotification, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID, limit, offset)
	}

	return nil, nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.markAsReadFunc != nil {
		return m.markAsReadFunc(ctx, id)
	}

	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.markAllAsReadFunc != nil {
		return m.markAllAsReadFunc(ctx, userID)
	}

	return nil
}

func (m *mockNotificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.getUnreadCountFunc != nil {
		return m.getUnreadCountFunc(ctx, userID)
	}

	return 0, nil
}

func TestInAppUseCase_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    *mockNotificationRepo
		input   *notification.InAppNotification
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockNotificationRepo{},
			input: &notification.InAppNotification{
				UserID: uuid.New(),
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockNotificationRepo{
				storeFunc: func(_ context.Context, _ *notification.InAppNotification) error {
					return errRepo
				},
			},
			input: &notification.InAppNotification{
				UserID: uuid.New(),
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewInAppUseCase(tt.repo)
			err := uc.Create(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestInAppUseCase_GetByID(t *testing.T) {
	t.Parallel()

	notificationID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockNotificationRepo
		id      uuid.UUID
		want    *notification.InAppNotification
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockNotificationRepo{
				getByIDFunc: func(_ context.Context, _ uuid.UUID) (*notification.InAppNotification, error) {
					return &notification.InAppNotification{
						ID:     notificationID,
						UserID: userID,
						Title:  "Test",
					}, nil
				},
			},
			id: notificationID,
			want: &notification.InAppNotification{
				ID:     notificationID,
				UserID: userID,
				Title:  "Test",
			},
			wantErr: false,
		},
		{
			name: "not found",
			repo: &mockNotificationRepo{
				getByIDFunc: func(_ context.Context, _ uuid.UUID) (*notification.InAppNotification, error) {
					return nil, errRepo
				},
			},
			id:      notificationID,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewInAppUseCase(tt.repo)
			got, err := uc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInAppUseCase_GetUnreadCount(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockNotificationRepo
		userID  uuid.UUID
		want    int
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockNotificationRepo{
				getUnreadCountFunc: func(_ context.Context, _ uuid.UUID) (int, error) {
					return 5, nil
				},
			},
			userID:  userID,
			want:    5,
			wantErr: false,
		},
		{
			name: "repo error",
			repo: &mockNotificationRepo{
				getUnreadCountFunc: func(_ context.Context, _ uuid.UUID) (int, error) {
					return 0, errRepo
				},
			},
			userID:  userID,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := notificationuc.NewInAppUseCase(tt.repo)
			got, err := uc.GetUnreadCount(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
