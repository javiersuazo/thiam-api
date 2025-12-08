package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
)

const defaultSMTPTimeout = 30 * time.Second

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
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = strings.Join(msg.To, ", ")
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"

	var body string

	if msg.HTMLBody != "" {
		headers["Content-Type"] = "text/html; charset=UTF-8"
		body = msg.HTMLBody
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
		body = msg.Body
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + body

	if s.config.UseTLS {
		return s.sendWithTLS(ctx, addr, auth, msg.To, []byte(message))
	}

	return smtp.SendMail(addr, auth, s.config.From, msg.To, []byte(message))
}

func (s *SMTPSender) sendWithTLS(ctx context.Context, addr string, auth smtp.Auth, to []string, msg []byte) error {
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
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err = client.Mail(s.config.From); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return client.Quit()
}
