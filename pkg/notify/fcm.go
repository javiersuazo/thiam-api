package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
)

const (
	fcmEndpoint       = "https://fcm.googleapis.com/fcm/send"
	defaultFCMTimeout = 10 * time.Second
)

var errFCMBadStatus = errors.New("fcm returned non-200 status")

type FCMSendResult struct {
	SuccessCount int
	FailureCount int
	FailedTokens []FailedToken
}

type FailedToken struct {
	Token string
	Error string
}

type FCMConfig struct {
	ServerKey string
	Timeout   time.Duration
}

type FCMSender struct {
	config FCMConfig
	client *http.Client
}

func NewFCMSender(config FCMConfig) *FCMSender {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultFCMTimeout
	}

	return &FCMSender{
		config: config,
		client: &http.Client{Timeout: timeout},
	}
}

type fcmMessage struct {
	RegistrationIDs  []string          `json:"registration_ids,omitempty"`
	To               string            `json:"to,omitempty"`
	Notification     *fcmNotification  `json:"notification,omitempty"`
	Data             map[string]string `json:"data,omitempty"`
	Priority         string            `json:"priority,omitempty"`
	ContentAvailable bool              `json:"content_available,omitempty"`
}

type fcmNotification struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	ImageURL string `json:"image,omitempty"`
	Sound    string `json:"sound,omitempty"`
	Badge    string `json:"badge,omitempty"`
}

type fcmResponse struct {
	MulticastID int64       `json:"multicast_id"`
	Success     int         `json:"success"`
	Failure     int         `json:"failure"`
	Results     []fcmResult `json:"results"`
}

type fcmResult struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

func (f *FCMSender) Send(ctx context.Context, msg *notification.PushMessage, tokens []string) error {
	_, err := f.SendWithResult(ctx, msg, tokens)

	return err
}

func (f *FCMSender) SendWithResult(ctx context.Context, msg *notification.PushMessage, tokens []string) (*FCMSendResult, error) {
	if len(tokens) == 0 {
		return &FCMSendResult{}, nil
	}

	fcmMsg := fcmMessage{
		RegistrationIDs: tokens,
		Notification: &fcmNotification{
			Title:    msg.Title,
			Body:     msg.Body,
			ImageURL: msg.ImageURL,
			Sound:    msg.Sound,
		},
		Data:     msg.Data,
		Priority: "high",
	}

	payload, err := json.Marshal(fcmMsg)
	if err != nil {
		return nil, fmt.Errorf("marshal fcm message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fcmEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "key="+f.config.ServerKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errFCMBadStatus, resp.StatusCode)
	}

	var fcmResp fcmResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &FCMSendResult{
		SuccessCount: fcmResp.Success,
		FailureCount: fcmResp.Failure,
	}

	if fcmResp.Failure > 0 {
		result.FailedTokens = make([]FailedToken, 0, fcmResp.Failure)
		for i, r := range fcmResp.Results {
			if r.Error != "" {
				result.FailedTokens = append(result.FailedTokens, FailedToken{
					Token: tokens[i],
					Error: r.Error,
				})
			}
		}
	}

	return result, nil
}
