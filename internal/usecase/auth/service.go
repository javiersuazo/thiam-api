package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/internal/repo"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
	"github.com/google/uuid"
)

type Service struct {
	userRepo         repo.UserRepo
	refreshTokenRepo repo.RefreshTokenRepo
	jwtService       *pkgauth.JWTService
}

type ServiceDeps struct {
	UserRepo         repo.UserRepo
	RefreshTokenRepo repo.RefreshTokenRepo
	JWTService       *pkgauth.JWTService
}

func NewService(deps *ServiceDeps) *Service {
	return &Service{
		userRepo:         deps.UserRepo,
		refreshTokenRepo: deps.RefreshTokenRepo,
		jwtService:       deps.JWTService,
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
	SessionID uuid.UUID
}

func (s *Service) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email existence: %w", err)
	}

	if exists {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := pkgauth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &auth.User{
		Email:        input.Email,
		PasswordHash: &passwordHash,
		Name:         &input.Name,
		Status:       auth.UserStatusPendingVerify,
		AuthProvider: auth.ProviderLocal,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	sessionID := uuid.New()

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generating token pair: %w", err)
	}

	refreshToken := &auth.RefreshToken{
		UserID:    user.ID,
		TokenHash: pkgauth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenDuration()),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &RegisterOutput{
		User:      user,
		TokenPair: tokenPair,
		SessionID: sessionID,
	}, nil
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginOutput struct {
	User      *auth.User
	TokenPair *pkgauth.TokenPair
	SessionID uuid.UUID
}

func (s *Service) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.CanLogin() {
		if user.IsLocked() {
			return nil, ErrAccountLocked
		}

		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	valid, err := pkgauth.VerifyPassword(input.Password, *user.PasswordHash)
	if err != nil || !valid {
		s.handleFailedLogin(ctx, user)

		return nil, ErrInvalidCredentials
	}

	s.handleSuccessfulLogin(ctx, user, input.IPAddress)

	sessionID := uuid.New()

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generating token pair: %w", err)
	}

	refreshToken := &auth.RefreshToken{
		UserID:    user.ID,
		TokenHash: pkgauth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenDuration()),
		IPAddress: &input.IPAddress,
		UserAgent: &input.UserAgent,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &LoginOutput{
		User:      user,
		TokenPair: tokenPair,
		SessionID: sessionID,
	}, nil
}

func (s *Service) handleFailedLogin(ctx context.Context, user *auth.User) {
	user.IncrementFailedLogins()
	//nolint:errcheck // best-effort login tracking, should not fail main operation
	s.userRepo.Update(ctx, user)
}

func (s *Service) handleSuccessfulLogin(ctx context.Context, user *auth.User, ipAddress string) {
	user.RecordLogin(ipAddress)
	//nolint:errcheck // best-effort login tracking, should not fail main operation
	s.userRepo.Update(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := pkgauth.HashToken(refreshTokenStr)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if err := s.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("revoking token: %w", err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.refreshTokenRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("revoking all tokens: %w", err)
	}

	return nil
}

type RefreshTokenInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

type RefreshTokenOutput struct {
	TokenPair *pkgauth.TokenPair
	SessionID uuid.UUID
}

func (s *Service) RefreshToken(ctx context.Context, input *RefreshTokenInput) (*RefreshTokenOutput, error) {
	claims, err := s.jwtService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenHash := pkgauth.HashToken(input.RefreshToken)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !storedToken.IsValid() {
		return nil, ErrTokenRevoked
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.CanLogin() {
		return nil, ErrAccountLocked
	}

	if err := s.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	sessionID := uuid.New()

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, sessionID)
	if err != nil {
		return nil, fmt.Errorf("generating token pair: %w", err)
	}

	newRefreshToken := &auth.RefreshToken{
		UserID:    user.ID,
		TokenHash: pkgauth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenDuration()),
		IPAddress: &input.IPAddress,
		UserAgent: &input.UserAgent,
	}

	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &RefreshTokenOutput{
		TokenPair: tokenPair,
		SessionID: sessionID,
	}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*auth.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*pkgauth.Claims, error) {
	return s.jwtService.ValidateAccessToken(tokenString)
}
