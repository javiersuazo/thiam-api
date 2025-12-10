package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/internal/repo"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
)

const minPasswordLength = 8

type RegisterUseCase struct {
	userRepo         repo.UserRepo
	refreshTokenRepo repo.RefreshTokenRepo
	jwtService       *pkgauth.JWTService
}

func NewRegisterUseCase(
	userRepo repo.UserRepo,
	refreshTokenRepo repo.RefreshTokenRepo,
	jwtService *pkgauth.JWTService,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type RegisterOutput struct {
	User      *auth.User
	TokenPair *pkgauth.TokenPair
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := uc.validate(input); err != nil {
		return nil, err
	}

	exists, err := uc.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("RegisterUseCase.Execute - checking email: %w", err)
	}

	if exists {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := pkgauth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("RegisterUseCase.Execute - hashing password: %w", err)
	}

	user := &auth.User{
		Email:        input.Email,
		PasswordHash: &passwordHash,
		Status:       auth.UserStatusPendingVerification,
	}

	if input.Name != "" {
		name := strings.TrimSpace(input.Name)
		user.Name = &name
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("RegisterUseCase.Execute - creating user: %w", err)
	}

	tokenPair, refreshExpiresAt, err := uc.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("RegisterUseCase.Execute - generating tokens: %w", err)
	}

	refreshToken := &auth.RefreshToken{
		UserID:     user.ID,
		TokenHash:  pkgauth.HashToken(tokenPair.RefreshToken),
		Generation: 1,
		ExpiresAt:  refreshExpiresAt,
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("RegisterUseCase.Execute - storing refresh token: %w", err)
	}

	return &RegisterOutput{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

func (uc *RegisterUseCase) validate(input RegisterInput) error {
	if !strings.Contains(input.Email, "@") {
		return ErrInvalidEmail
	}

	if len(input.Password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}
