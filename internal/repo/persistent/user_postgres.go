package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errUserNotFound = errors.New("user not found")

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(ctx context.Context, user *auth.User) error {
	now := time.Now().UTC()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	user.CreatedAt = now
	user.UpdatedAt = now

	sql, args, err := r.Builder.
		Insert("users").
		Columns(
			"id", "email", "password_hash", "name", "avatar_url",
			"phone_number", "phone_verified", "phone_verified_at",
			"email_verified", "email_verified_at",
			"mfa_enabled", "mfa_secret", "mfa_recovery_codes",
			"status", "failed_login_attempts", "locked_until",
			"last_login_at", "last_login_ip",
			"auth_provider", "provider_user_id",
			"created_at", "updated_at",
		).
		Values(
			user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL,
			user.PhoneNumber, user.PhoneVerified, user.PhoneVerifiedAt,
			user.EmailVerified, user.EmailVerifiedAt,
			user.MFAEnabled, user.MFASecret, user.MFARecoveryCodes,
			user.Status, user.FailedLoginAttempts, user.LockedUntil,
			user.LastLoginAt, user.LastLoginIP,
			user.AuthProvider, user.ProviderUserID,
			user.CreatedAt, user.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	sql, args, err := r.Builder.
		Select(userColumns()...).
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByID - r.Builder: %w", err)
	}

	return r.scanUser(ctx, sql, args)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	sql, args, err := r.Builder.
		Select(userColumns()...).
		From("users").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetByEmail - r.Builder: %w", err)
	}

	return r.scanUser(ctx, sql, args)
}

func (r *UserRepo) Update(ctx context.Context, user *auth.User) error {
	user.UpdatedAt = time.Now().UTC()

	sql, args, err := r.Builder.
		Update("users").
		Set("email", user.Email).
		Set("password_hash", user.PasswordHash).
		Set("name", user.Name).
		Set("avatar_url", user.AvatarURL).
		Set("phone_number", user.PhoneNumber).
		Set("phone_verified", user.PhoneVerified).
		Set("phone_verified_at", user.PhoneVerifiedAt).
		Set("email_verified", user.EmailVerified).
		Set("email_verified_at", user.EmailVerifiedAt).
		Set("mfa_enabled", user.MFAEnabled).
		Set("mfa_secret", user.MFASecret).
		Set("mfa_recovery_codes", user.MFARecoveryCodes).
		Set("status", user.Status).
		Set("failed_login_attempts", user.FailedLoginAttempts).
		Set("locked_until", user.LockedUntil).
		Set("last_login_at", user.LastLoginAt).
		Set("last_login_ip", user.LastLoginIP).
		Set("auth_provider", user.AuthProvider).
		Set("provider_user_id", user.ProviderUserID).
		Set("updated_at", user.UpdatedAt).
		Where("id = ?", user.ID).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Update - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Update - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.Builder.
		Delete("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Delete - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Delete - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	sql, args, err := r.Builder.
		Select("1").
		From("users").
		Where("email = ?", email).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("UserRepo - ExistsByEmail - r.Builder: %w", err)
	}

	var exists int

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("UserRepo - ExistsByEmail - r.Pool.QueryRow: %w", err)
	}

	return true, nil
}

func (r *UserRepo) scanUser(ctx context.Context, sql string, args []interface{}) (*auth.User, error) {
	var user auth.User

	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL,
		&user.PhoneNumber, &user.PhoneVerified, &user.PhoneVerifiedAt,
		&user.EmailVerified, &user.EmailVerifiedAt,
		&user.MFAEnabled, &user.MFASecret, &user.MFARecoveryCodes,
		&user.Status, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.LastLoginAt, &user.LastLoginIP,
		&user.AuthProvider, &user.ProviderUserID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errUserNotFound
		}

		return nil, fmt.Errorf("UserRepo - scanUser - r.Pool.QueryRow: %w", err)
	}

	return &user, nil
}

func userColumns() []string {
	return []string{
		"id", "email", "password_hash", "name", "avatar_url",
		"phone_number", "phone_verified", "phone_verified_at",
		"email_verified", "email_verified_at",
		"mfa_enabled", "mfa_secret", "mfa_recovery_codes",
		"status", "failed_login_attempts", "locked_until",
		"last_login_at", "last_login_ip",
		"auth_provider", "provider_user_id",
		"created_at", "updated_at",
	}
}
