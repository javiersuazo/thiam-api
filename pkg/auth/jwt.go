package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT validation and generation errors.
var (
	// ErrInvalidToken is returned when the token cannot be parsed or is malformed.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token has passed its expiration time.
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidClaims is returned when the token claims cannot be extracted or are invalid.
	ErrInvalidClaims = errors.New("invalid token claims")
	// ErrMissingSecret is returned when no JWT secret is provided to the service.
	ErrMissingSecret = errors.New("jwt secret is required")
	// ErrSecretTooShort is returned when the JWT secret is less than 32 characters.
	ErrSecretTooShort = errors.New("jwt secret must be at least 32 characters for HS256")
	// ErrUnexpectedSigning is returned when the token uses a signing method other than HMAC.
	ErrUnexpectedSigning = errors.New("unexpected signing method")
)

// TokenType represents the type of JWT token (access or refresh).
type TokenType string

// Token type constants.
const (
	// TokenTypeAccess identifies short-lived tokens used for API authentication.
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh identifies long-lived tokens used to obtain new access tokens.
	TokenTypeRefresh TokenType = "refresh"
)

const (
	defaultAccessDuration  = 15 * time.Minute
	defaultRefreshDuration = 7 * 24 * time.Hour
	defaultIssuer          = "thiam"
	minSecretLength        = 32
)

// Claims represents the custom JWT claims including user information and token type.
type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
}

// TokenPair contains the access and refresh tokens returned after authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// JWTConfig holds the configuration for the JWT service.
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// JWTService handles JWT token generation and validation using HS256 signing.
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service with the provided configuration.
// Returns ErrMissingSecret if no secret is provided, or ErrSecretTooShort if the secret
// is less than 32 characters. Default durations are 15 minutes for access tokens
// and 7 days for refresh tokens.
func NewJWTService(config JWTConfig) (*JWTService, error) {
	if config.Secret == "" {
		return nil, ErrMissingSecret
	}

	if len(config.Secret) < minSecretLength {
		return nil, ErrSecretTooShort
	}

	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = defaultAccessDuration
	}

	if config.RefreshTokenDuration == 0 {
		config.RefreshTokenDuration = defaultRefreshDuration
	}

	if config.Issuer == "" {
		config.Issuer = defaultIssuer
	}

	return &JWTService{config: config}, nil
}

// GenerateTokenPair creates a new access and refresh token pair for the given user.
// Returns the token pair, the refresh token expiration time, and any error encountered.
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, time.Time, error) {
	accessToken, err := s.generateToken(userID, email, TokenTypeAccess, s.config.AccessTokenDuration)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.generateToken(userID, email, TokenTypeRefresh, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}

	refreshExpiresAt := time.Now().Add(s.config.RefreshTokenDuration)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	}, refreshExpiresAt, nil
}

func (s *JWTService) generateToken(userID uuid.UUID, email string, tokenType TokenType, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string, returning the claims if valid.
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

// ValidateAccessToken validates a token and ensures it is an access token.
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

// ValidateRefreshToken validates a token and ensures it is a refresh token.
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

func (s *JWTService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigning
	}

	return []byte(s.config.Secret), nil
}

// GetRefreshTokenDuration returns the configured refresh token duration.
func (s *JWTService) GetRefreshTokenDuration() time.Duration {
	return s.config.RefreshTokenDuration
}
