package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
)

const defaultSMTPTimeout = 30 * time.Second

var errTLSRequired = errors.New("TLS is required for secure email delivery")

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type SMTPSender struct {
	config SMTPConfig
}

func NewSMTPSender(config *SMTPConfig) *SMTPSender {
	return &SMTPSender{config: *config}
}

func (s *SMTPSender) Send(ctx context.Context, msg *notification.EmailMessage) error {
	if !s.config.UseTLS {
		return errTLSRequired
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	message := s.buildMessage(msg)

	return s.sendWithTLS(ctx, addr, auth, msg.To, message)
}

func (s *SMTPSender) buildMessage(msg *notification.EmailMessage) []byte {
	contentType := "text/plain; charset=UTF-8"
	body := msg.Body

	if msg.HTMLBody != "" {
		contentType = "text/html; charset=UTF-8"
		body = msg.HTMLBody
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	sb.WriteString("\r\n")
	sb.WriteString(body)

	return []byte(sb.String())
}

func (s *SMTPSender) sendWithTLS(ctx context.Context, addr string, auth smtp.Auth, to []string, msg []byte) error {
	client, err := s.dialTLS(ctx, addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if authErr := s.authenticate(client, auth); authErr != nil {
		return authErr
	}

	return s.sendMessage(client, to, msg)
}

func (s *SMTPSender) dialTLS(ctx context.Context, addr string) (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: defaultSMTPTimeout},
		Config:    tlsConfig,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf("smtp client: %w", err)
	}

	return client, nil
}

func (s *SMTPSender) authenticate(client *smtp.Client, auth smtp.Auth) error {
	if auth == nil {
		return nil
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	return nil
}

func (s *SMTPSender) sendMessage(client *smtp.Client, to []string, msg []byte) error {
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return client.Quit()
}
