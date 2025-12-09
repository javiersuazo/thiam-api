package persistent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotificationRepo(t *testing.T) {
	t.Parallel()

	repo := NewNotificationRepo(nil)
	assert.NotNil(t, repo)
}

func TestNewNotificationPreferencesRepo(t *testing.T) {
	t.Parallel()

	repo := NewNotificationPreferencesRepo(nil)
	assert.NotNil(t, repo)
}

func TestNewPushTokenRepo(t *testing.T) {
	t.Parallel()

	repo := NewPushTokenRepo(nil)
	assert.NotNil(t, repo)
}

func TestNewDeliveryLogRepo(t *testing.T) {
	t.Parallel()

	repo := NewDeliveryLogRepo(nil)
	assert.NotNil(t, repo)
}

func TestNewOutboxRepo(t *testing.T) {
	t.Parallel()

	repo := NewOutboxRepo(nil)
	assert.NotNil(t, repo)
}
