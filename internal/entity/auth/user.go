// Package auth contains authentication domain entities.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the current state of a user account.
type UserStatus string

// User status constants.
const (
	UserStatusActive          UserStatus = "active"
	UserStatusInactive        UserStatus = "inactive"
	UserStatusLocked          UserStatus = "locked"
	UserStatusPendingVerify   UserStatus = "pending_verification"
	UserStatusSuspended       UserStatus = "suspended"
	UserStatusDeleted         UserStatus = "deleted"
	UserStatusPasswordReset   UserStatus = "password_reset_required"
	UserStatusPendingMFASetup UserStatus = "pending_mfa_setup"
)

// Provider represents the authentication provider.
type Provider string

// Authentication provider constants.
const (
	ProviderLocal    Provider = "local"
	ProviderGoogle   Provider = "google"
	ProviderApple    Provider = "apple"
	ProviderFacebook Provider = "facebook"
	ProviderGitHub   Provider = "github"
)

// User represents a user in the system.
type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        *string    `json:"-"`
	Name                *string    `json:"name,omitempty"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
	PhoneNumber         *string    `json:"phone_number,omitempty"`
	PhoneVerified       bool       `json:"phone_verified"`
	PhoneVerifiedAt     *time.Time `json:"phone_verified_at,omitempty"`
	EmailVerified       bool       `json:"email_verified"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at,omitempty"`
	MFAEnabled          bool       `json:"mfa_enabled"`
	MFASecret           *string    `json:"-"`
	MFARecoveryCodes    []string   `json:"-"`
	Status              UserStatus `json:"status"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty"`
	AuthProvider        Provider   `json:"auth_provider"`
	ProviderUserID      *string    `json:"provider_user_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// IsLocked returns true if the user account is currently locked.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}

	return time.Now().Before(*u.LockedUntil)
}

// CanLogin returns true if the user is allowed to login.
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive && !u.IsLocked()
}

// RequiresMFA returns true if the user has MFA enabled and configured.
func (u *User) RequiresMFA() bool {
	return u.MFAEnabled && u.MFASecret != nil
}

// IncrementFailedLogins increases the failed login counter.
func (u *User) IncrementFailedLogins() {
	u.FailedLoginAttempts++
}

// ResetFailedLogins clears the failed login counter and unlock time.
func (u *User) ResetFailedLogins() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
}

// Lock sets the user account as locked for the specified duration.
func (u *User) Lock(duration time.Duration) {
	until := time.Now().Add(duration)
	u.LockedUntil = &until
	u.Status = UserStatusLocked
}

// MarkEmailVerified marks the user's email as verified.
func (u *User) MarkEmailVerified() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now

	if u.Status == UserStatusPendingVerify {
		u.Status = UserStatusActive
	}
}

// MarkPhoneVerified marks the user's phone as verified.
func (u *User) MarkPhoneVerified() {
	now := time.Now()
	u.PhoneVerified = true
	u.PhoneVerifiedAt = &now
}

// RecordLogin updates the user's last login information.
func (u *User) RecordLogin(ip string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = &ip
	u.ResetFailedLogins()
}
