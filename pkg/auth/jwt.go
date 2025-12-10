package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT errors.
var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidClaims     = errors.New("invalid token claims")
	ErrMissingSecret     = errors.New("jwt secret is required")
	ErrUnexpectedSigning = errors.New("unexpected signing method")
)

// TokenType identifies the type of JWT token.
type TokenType string

// Token type constants.
const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Default token durations.
const (
	DefaultAccessDuration  = 15 * time.Minute
	DefaultRefreshDuration = 24 * time.Hour
	DefaultIssuer          = "thiam"
)

// Claims represents the JWT claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	SessionID uuid.UUID `json:"session_id,omitempty"`
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// JWTService handles JWT token operations.
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service.
func NewJWTService(config JWTConfig) (*JWTService, error) {
	if config.Secret == "" {
		return nil, ErrMissingSecret
	}

	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = DefaultAccessDuration
	}

	if config.RefreshTokenDuration == 0 {
		config.RefreshTokenDuration = DefaultRefreshDuration
	}

	if config.Issuer == "" {
		config.Issuer = DefaultIssuer
	}

	return &JWTService{config: config}, nil
}

// GenerateAccessToken creates a new access token.
func (s *JWTService) GenerateAccessToken(userID uuid.UUID, email string, sessionID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.config.AccessTokenDuration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Email:     email,
		TokenType: TokenTypeAccess,
		SessionID: sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a new refresh token.
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.config.RefreshTokenDuration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Email:     email,
		TokenType: TokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateTokenPair creates both access and refresh tokens.
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string, sessionID uuid.UUID) (*TokenPair, error) {
	accessToken, expiresAt, err := s.GenerateAccessToken(userID, email, sessionID)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.GenerateRefreshToken(userID, email)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken parses and validates any token type.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, s.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func (s *JWTService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigning
	}

	return []byte(s.config.Secret), nil
}

// ValidateAccessToken validates an access token specifically.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token specifically.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessTokenDuration returns the configured access token duration.
func (s *JWTService) GetAccessTokenDuration() time.Duration {
	return s.config.AccessTokenDuration
}

// GetRefreshTokenDuration returns the configured refresh token duration.
func (s *JWTService) GetRefreshTokenDuration() time.Duration {
	return s.config.RefreshTokenDuration
}
