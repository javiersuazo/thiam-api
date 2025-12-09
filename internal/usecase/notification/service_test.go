package notification_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	notificationuc "github.com/evrone/go-clean-template/internal/usecase/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockDeliveryLogRepo struct {
	storeFunc               func(ctx context.Context, log *notification.DeliveryLog) error
	getByNotificationIDFunc func(ctx context.Context, notificationID uuid.UUID) ([]notification.DeliveryLog, error)
}

func (m *mockDeliveryLogRepo) Store(ctx context.Context, log *notification.DeliveryLog) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, log)
	}

	return nil
}

func (m *mockDeliveryLogRepo) GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]notification.DeliveryLog, error) {
	if m.getByNotificationIDFunc != nil {
		return m.getByNotificationIDFunc(ctx, notificationID)
	}

	return nil, nil
}

type mockEmailSender struct {
	sendFunc func(ctx context.Context, msg *notification.EmailMessage) error
}

func (m *mockEmailSender) Send(ctx context.Context, msg *notification.EmailMessage) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg)
	}

	return nil
}

type mockPushSender struct {
	sendFunc func(ctx context.Context, msg *notification.PushMessage, tokens []string) error
}

func (m *mockPushSender) Send(ctx context.Context, msg *notification.PushMessage, tokens []string) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg, tokens)
	}

	return nil
}

func TestService_SendInApp(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name        string
		notifRepo   *mockNotificationRepo
		prefsRepo   *mockPreferencesRepo
		deliveryLog *mockDeliveryLogRepo
		input       *notification.InAppMessage
		wantErr     bool
	}{
		{
			name:      "success",
			notifRepo: &mockNotificationRepo{},
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{InAppEnabled: true}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			input: &notification.InAppMessage{
				UserID: userID,
				Type:   "test",
				Title:  "Test Title",
				Body:   "Test Body",
			},
			wantErr: false,
		},
		{
			name:      "disabled by preferences",
			notifRepo: &mockNotificationRepo{},
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{InAppEnabled: false}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			input: &notification.InAppMessage{
				UserID: userID,
				Type:   "test",
				Title:  "Test Title",
				Body:   "Test Body",
			},
			wantErr: false,
		},
		{
			name: "repo error",
			notifRepo: &mockNotificationRepo{
				storeFunc: func(_ context.Context, _ *notification.InAppNotification) error {
					return errRepo
				},
			},
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return nil, errRepo
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			input: &notification.InAppMessage{
				UserID: userID,
				Type:   "test",
				Title:  "Test Title",
				Body:   "Test Body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := notificationuc.NewService(&notificationuc.ServiceDeps{
				NotificationRepo: tt.notifRepo,
				PrefsRepo:        tt.prefsRepo,
				PushTokenRepo:    &mockPushTokenRepo{},
				DeliveryLogRepo:  tt.deliveryLog,
			})

			err := svc.SendInApp(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_SendPush(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name          string
		prefsRepo     *mockPreferencesRepo
		pushTokenRepo *mockPushTokenRepo
		deliveryLog   *mockDeliveryLogRepo
		pushSender    *mockPushSender
		input         *notification.PushMessage
		wantErr       bool
	}{
		{
			name: "success",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{PushEnabled: true}, nil
				},
			},
			pushTokenRepo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{
						{Token: "token-1", Active: true},
						{Token: "token-2", Active: true},
					}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			pushSender:  &mockPushSender{},
			input: &notification.PushMessage{
				UserID: userID,
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: false,
		},
		{
			name: "disabled by preferences",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{PushEnabled: false}, nil
				},
			},
			pushTokenRepo: &mockPushTokenRepo{},
			deliveryLog:   &mockDeliveryLogRepo{},
			pushSender:    &mockPushSender{},
			input: &notification.PushMessage{
				UserID: userID,
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: false,
		},
		{
			name: "no tokens",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{PushEnabled: true}, nil
				},
			},
			pushTokenRepo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			pushSender:  &mockPushSender{},
			input: &notification.PushMessage{
				UserID: userID,
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: false,
		},
		{
			name: "no active tokens",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{PushEnabled: true}, nil
				},
			},
			pushTokenRepo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{
						{Token: "token-1", Active: false},
					}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			pushSender:  &mockPushSender{},
			input: &notification.PushMessage{
				UserID: userID,
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: false,
		},
		{
			name: "send error",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{PushEnabled: true}, nil
				},
			},
			pushTokenRepo: &mockPushTokenRepo{
				getByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]notification.PushToken, error) {
					return []notification.PushToken{{Token: "token-1", Active: true}}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			pushSender: &mockPushSender{
				sendFunc: func(_ context.Context, _ *notification.PushMessage, _ []string) error {
					return errRepo
				},
			},
			input: &notification.PushMessage{
				UserID: userID,
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := notificationuc.NewService(&notificationuc.ServiceDeps{
				NotificationRepo: &mockNotificationRepo{},
				PrefsRepo:        tt.prefsRepo,
				PushTokenRepo:    tt.pushTokenRepo,
				DeliveryLogRepo:  tt.deliveryLog,
				PushSender:       tt.pushSender,
			})

			err := svc.SendPush(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_SendEmail(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name        string
		prefsRepo   *mockPreferencesRepo
		deliveryLog *mockDeliveryLogRepo
		emailSender *mockEmailSender
		input       *notification.EmailMessage
		wantErr     bool
	}{
		{
			name: "success",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{EmailEnabled: true}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			emailSender: &mockEmailSender{},
			input: &notification.EmailMessage{
				UserID:  userID,
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr: false,
		},
		{
			name: "disabled by preferences",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return &notification.UserPreferences{EmailEnabled: false}, nil
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			emailSender: &mockEmailSender{},
			input: &notification.EmailMessage{
				UserID:  userID,
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr: false,
		},
		{
			name: "send error",
			prefsRepo: &mockPreferencesRepo{
				getFunc: func(_ context.Context, _ uuid.UUID) (*notification.UserPreferences, error) {
					return nil, errRepo
				},
			},
			deliveryLog: &mockDeliveryLogRepo{},
			emailSender: &mockEmailSender{
				sendFunc: func(_ context.Context, _ *notification.EmailMessage) error {
					return errRepo
				},
			},
			input: &notification.EmailMessage{
				UserID:  userID,
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := notificationuc.NewService(&notificationuc.ServiceDeps{
				NotificationRepo: &mockNotificationRepo{},
				PrefsRepo:        tt.prefsRepo,
				PushTokenRepo:    &mockPushTokenRepo{},
				DeliveryLogRepo:  tt.deliveryLog,
				EmailSender:      tt.emailSender,
			})

			err := svc.SendEmail(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
