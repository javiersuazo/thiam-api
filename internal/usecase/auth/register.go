// Package auth provides authentication use cases for user registration, login, and token management.
package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/internal/repo"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
)

// Registration validation errors.
var (
	// ErrEmailAlreadyExists is returned when attempting to register with an email that is already in use.
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidEmail is returned when the email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrPasswordTooShort is returned when the password is less than 8 characters.
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrPasswordTooWeak is returned when the password doesn't meet complexity requirements.
	ErrPasswordTooWeak = errors.New("password must contain uppercase, lowercase, digit, and special character")
)

const minPasswordLength = 8

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// RegisterUseCase handles user registration with email and password.
type RegisterUseCase struct {
	userRepo         repo.UserRepo
	refreshTokenRepo repo.RefreshTokenRepo
	jwtService       *pkgauth.JWTService
}

// NewRegisterUseCase creates a new RegisterUseCase with the required dependencies.
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

// RegisterInput contains the data required to register a new user.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// RegisterOutput contains the registered user and authentication tokens.
type RegisterOutput struct {
	User      *auth.User
	TokenPair *pkgauth.TokenPair
}

// Execute registers a new user with the provided email and password.
// It normalizes the email, validates input, hashes the password, creates the user,
// and generates authentication tokens.
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
	if !emailRegex.MatchString(input.Email) {
		return ErrInvalidEmail
	}

	if len(input.Password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	if !isPasswordComplex(input.Password) {
		return ErrPasswordTooWeak
	}

	return nil
}

func isPasswordComplex(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}
