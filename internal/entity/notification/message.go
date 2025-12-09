package notification

import "github.com/google/uuid"

type EmailMessage struct {
	UserID      uuid.UUID
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type SMSMessage struct {
	To   string
	Body string
}

type PushMessage struct {
	UserID   uuid.UUID
	Title    string
	Body     string
	Data     map[string]string
	ImageURL string
	Badge    *int
	Sound    string
}

type InAppMessage struct {
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      map[string]string
	ActionURL string
	ImageURL  string
}
