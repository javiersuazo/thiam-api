package notification

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
	ChannelInApp Channel = "in_app"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusSent      Status = "sent"
	StatusFailed    Status = "failed"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

type Notification struct {
	ID          uuid.UUID         `json:"id"`
	UserID      uuid.UUID         `json:"user_id"`
	Channel     Channel           `json:"channel"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
	Priority    Priority          `json:"priority"`
	Status      Status            `json:"status"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	SentAt      *time.Time        `json:"sent_at,omitempty"`
	ReadAt      *time.Time        `json:"read_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type InAppNotification struct {
	ID        uuid.UUID         `json:"id"`
	UserID    uuid.UUID         `json:"user_id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Data      map[string]string `json:"data,omitempty"`
	ActionURL *string           `json:"action_url,omitempty"`
	ImageURL  *string           `json:"image_url,omitempty"`
	Read      bool              `json:"read"`
	ReadAt    *time.Time        `json:"read_at,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type UserPreferences struct {
	UserID       uuid.UUID `json:"user_id"`
	EmailEnabled bool      `json:"email_enabled"`
	SMSEnabled   bool      `json:"sms_enabled"`
	PushEnabled  bool      `json:"push_enabled"`
	InAppEnabled bool      `json:"in_app_enabled"`
	QuietStart   *string   `json:"quiet_start,omitempty"`
	QuietEnd     *string   `json:"quiet_end,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DeliveryLog struct {
	ID             uuid.UUID  `json:"id"`
	NotificationID uuid.UUID  `json:"notification_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Channel        Channel    `json:"channel"`
	Status         Status     `json:"status"`
	Provider       string     `json:"provider"`
	ProviderMsgID  *string    `json:"provider_message_id,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	Attempts       int        `json:"attempts"`
	CreatedAt      time.Time  `json:"created_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
}

type PushToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	DeviceID  string    `json:"device_id"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
