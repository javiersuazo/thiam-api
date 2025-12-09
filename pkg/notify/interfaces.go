package notify

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity/notification"
)

type EmailSender interface {
	Send(ctx context.Context, msg *notification.EmailMessage) error
}

type SMSSender interface {
	Send(ctx context.Context, msg *notification.SMSMessage) error
}

type PushSender interface {
	Send(ctx context.Context, msg *notification.PushMessage, tokens []string) error
}

type Notifier interface {
	SendEmail(ctx context.Context, msg *notification.EmailMessage) error
	SendSMS(ctx context.Context, msg *notification.SMSMessage) error
	SendPush(ctx context.Context, msg *notification.PushMessage) error
	SendInApp(ctx context.Context, msg *notification.InAppMessage) error
}
