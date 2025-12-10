package auth

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusDisabled            UserStatus = "disabled"
	UserStatusDeleted             UserStatus = "deleted"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        *string    `json:"-"`
	Name                *string    `json:"name,omitempty"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
	EmailVerified       bool       `json:"email_verified"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at,omitempty"`
	PhoneNumber         *string    `json:"phone_number,omitempty"`
	PhoneVerified       bool       `json:"phone_verified"`
	PhoneVerifiedAt     *time.Time `json:"phone_verified_at,omitempty"`
	Status              UserStatus `json:"status"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}

	return time.Now().Before(*u.LockedUntil)
}

func (u *User) CanLogin() bool {
	if u.IsLocked() {
		return false
	}

	return u.Status == UserStatusActive || u.Status == UserStatusPendingVerification
}
