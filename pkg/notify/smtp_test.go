package notify

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPSender(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	sender := NewSMTPSender(config)

	require.NotNil(t, sender)
	assert.Equal(t, config.Host, sender.config.Host)
	assert.Equal(t, config.Port, sender.config.Port)
	assert.Equal(t, config.From, sender.config.From)
	assert.True(t, sender.config.UseTLS)
}

func TestSMTPSender_Send_TLSRequired(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		Host:   "smtp.example.com",
		Port:   587,
		UseTLS: false,
	}

	sender := NewSMTPSender(config)
	msg := &notification.EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Test",
		Body:    "Test body",
	}

	err := sender.Send(context.Background(), msg)

	require.Error(t, err)
	assert.ErrorIs(t, err, errTLSRequired)
}

func TestSMTPSender_buildMessage_PlainText(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		From: "sender@example.com",
	}
	sender := NewSMTPSender(config)

	msg := &notification.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Plain text body",
	}

	result := sender.buildMessage(msg)
	resultStr := string(result)

	assert.Contains(t, resultStr, "From: sender@example.com\r\n")
	assert.Contains(t, resultStr, "To: recipient@example.com\r\n")
	assert.Contains(t, resultStr, "Subject: Test Subject\r\n")
	assert.Contains(t, resultStr, "MIME-Version: 1.0\r\n")
	assert.Contains(t, resultStr, "Content-Type: text/plain; charset=UTF-8\r\n")
	assert.Contains(t, resultStr, "Plain text body")
}

func TestSMTPSender_buildMessage_HTML(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		From: "sender@example.com",
	}
	sender := NewSMTPSender(config)

	msg := &notification.EmailMessage{
		To:       []string{"recipient@example.com"},
		Subject:  "Test Subject",
		HTMLBody: "<html><body>HTML body</body></html>",
	}

	result := sender.buildMessage(msg)
	resultStr := string(result)

	assert.Contains(t, resultStr, "Content-Type: text/html; charset=UTF-8\r\n")
	assert.Contains(t, resultStr, "<html><body>HTML body</body></html>")
}

func TestSMTPSender_buildMessage_MultipleRecipients(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		From: "sender@example.com",
	}
	sender := NewSMTPSender(config)

	msg := &notification.EmailMessage{
		To:      []string{"user1@example.com", "user2@example.com"},
		Subject: "Test Subject",
		Body:    "Body",
	}

	result := sender.buildMessage(msg)
	resultStr := string(result)

	assert.Contains(t, resultStr, "To: user1@example.com, user2@example.com\r\n")
}

func TestSMTPSender_buildMessage_HeaderOrder(t *testing.T) {
	t.Parallel()

	config := &SMTPConfig{
		From: "sender@example.com",
	}
	sender := NewSMTPSender(config)

	msg := &notification.EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Body",
	}

	result1 := sender.buildMessage(msg)
	result2 := sender.buildMessage(msg)

	assert.Equal(t, result1, result2, "Header order should be deterministic")
}
