package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/notification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFCMSender(t *testing.T) {
	t.Parallel()

	config := FCMConfig{
		ServerKey: "test-key",
		Timeout:   5 * time.Second,
	}

	sender := NewFCMSender(config)

	require.NotNil(t, sender)
	assert.Equal(t, config.ServerKey, sender.config.ServerKey)
	assert.Equal(t, 5*time.Second, sender.client.Timeout)
}

func TestNewFCMSender_DefaultTimeout(t *testing.T) {
	t.Parallel()

	config := FCMConfig{
		ServerKey: "test-key",
	}

	sender := NewFCMSender(config)

	assert.Equal(t, defaultFCMTimeout, sender.client.Timeout)
}

func TestFCMSender_Send_EmptyTokens(t *testing.T) {
	t.Parallel()

	sender := NewFCMSender(FCMConfig{ServerKey: "test-key"})
	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test",
		Body:   "Body",
	}

	err := sender.Send(context.Background(), msg, []string{})

	require.NoError(t, err)
}

func TestFCMSender_Send_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "key=test-key", r.Header.Get("Authorization"))

		resp := fcmResponse{
			MulticastID: 123,
			Success:     2,
			Failure:     0,
			Results: []fcmResult{
				{MessageID: "msg-1"},
				{MessageID: "msg-2"},
			},
		}

		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sender := &FCMSender{
		config: FCMConfig{ServerKey: "test-key"},
		client: server.Client(),
	}
	sender.client.Transport = &testTransport{server.URL}

	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test Title",
		Body:   "Test Body",
		Data:   map[string]string{"key": "value"},
	}

	err := sender.Send(context.Background(), msg, []string{"token-1", "token-2"})

	require.NoError(t, err)
}

func TestFCMSender_SendWithResult_PartialFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := fcmResponse{
			MulticastID: 123,
			Success:     1,
			Failure:     1,
			Results: []fcmResult{
				{MessageID: "msg-1"},
				{Error: "InvalidRegistration"},
			},
		}

		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // test helper
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sender := &FCMSender{
		config: FCMConfig{ServerKey: "test-key"},
		client: server.Client(),
	}
	sender.client.Transport = &testTransport{server.URL}

	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test",
		Body:   "Body",
	}

	result, err := sender.SendWithResult(context.Background(), msg, []string{"token-1", "token-2"})

	require.NoError(t, err)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
	require.Len(t, result.FailedTokens, 1)
	assert.Equal(t, "token-2", result.FailedTokens[0].Token)
	assert.Equal(t, "InvalidRegistration", result.FailedTokens[0].Error)
}

func TestFCMSender_Send_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	sender := &FCMSender{
		config: FCMConfig{ServerKey: "invalid-key"},
		client: server.Client(),
	}
	sender.client.Transport = &testTransport{server.URL}

	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test",
		Body:   "Body",
	}

	err := sender.Send(context.Background(), msg, []string{"token-1"})

	require.Error(t, err)
	assert.ErrorIs(t, err, errFCMBadStatus)
}

func TestFCMSender_Send_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // test helper
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	sender := &FCMSender{
		config: FCMConfig{ServerKey: "test-key"},
		client: server.Client(),
	}
	sender.client.Transport = &testTransport{server.URL}

	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test",
		Body:   "Body",
	}

	err := sender.Send(context.Background(), msg, []string{"token-1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestFCMSender_SendWithResult_EmptyTokens(t *testing.T) {
	t.Parallel()

	sender := NewFCMSender(FCMConfig{ServerKey: "test-key"})
	msg := &notification.PushMessage{
		UserID: uuid.New(),
		Title:  "Test",
		Body:   "Body",
	}

	result, err := sender.SendWithResult(context.Background(), msg, []string{})

	require.NoError(t, err)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Nil(t, result.FailedTokens)
}

type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[7:]

	return http.DefaultTransport.RoundTrip(req)
}
