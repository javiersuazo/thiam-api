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

var errRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepo struct {
	*postgres.Postgres
}

func NewRefreshTokenRepo(pg *postgres.Postgres) *RefreshTokenRepo {
	return &RefreshTokenRepo{pg}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, token *auth.RefreshToken) error {
	now := time.Now().UTC()

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	if token.FamilyID == uuid.Nil {
		token.FamilyID = uuid.New()
	}

	token.CreatedAt = now

	sql, args, err := r.Builder.
		Insert("refresh_tokens").
		Columns(
			"id", "user_id", "token_hash", "family_id", "generation",
			"device_info", "ip_address", "user_agent",
			"expires_at", "revoked_at", "last_used_at", "created_at",
		).
		Values(
			token.ID, token.UserID, token.TokenHash, token.FamilyID, token.Generation,
			token.DeviceInfo, token.IPAddress, token.UserAgent,
			token.ExpiresAt, token.RevokedAt, token.LastUsedAt, token.CreatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	sql, args, err := r.Builder.
		Select(
			"id", "user_id", "token_hash", "family_id", "generation",
			"device_info", "ip_address", "user_agent",
			"expires_at", "revoked_at", "last_used_at", "created_at",
		).
		From("refresh_tokens").
		Where("token_hash = ?", tokenHash).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("RefreshTokenRepo.GetByTokenHash - r.Builder: %w", err)
	}

	var token auth.RefreshToken

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.FamilyID, &token.Generation,
		&token.DeviceInfo, &token.IPAddress, &token.UserAgent,
		&token.ExpiresAt, &token.RevokedAt, &token.LastUsedAt, &token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errRefreshTokenNotFound
		}

		return nil, fmt.Errorf("RefreshTokenRepo.GetByTokenHash - r.Pool.QueryRow: %w", err)
	}

	return &token, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	sql, args, err := r.Builder.
		Update("refresh_tokens").
		Set("revoked_at", now).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.Revoke - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.Revoke - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID uuid.UUID) error {
	now := time.Now().UTC()

	sql, args, err := r.Builder.
		Update("refresh_tokens").
		Set("revoked_at", now).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		ToSql()
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.RevokeByFamilyID - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("RefreshTokenRepo.RevokeByFamilyID - r.Pool.Exec: %w", err)
	}

	return nil
}
