package auth

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token for session management.
type RefreshToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"`
	DeviceInfo *string    `json:"device_info,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsExpired returns true if the token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked returns true if the token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsValid returns true if the token is not expired and not revoked.
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}

// Revoke marks the token as revoked.
func (t *RefreshToken) Revoke() {
	now := time.Now()
	t.RevokedAt = &now
}

// MarkUsed updates the last used timestamp.
func (t *RefreshToken) MarkUsed() {
	now := time.Now()
	t.LastUsedAt = &now
}
