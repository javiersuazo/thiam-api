package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrInvalidClaims     = errors.New("invalid token claims")
	ErrMissingSecret     = errors.New("jwt secret is required")
	ErrUnexpectedSigning = errors.New("unexpected signing method")
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

const (
	defaultAccessDuration  = 15 * time.Minute
	defaultRefreshDuration = 7 * 24 * time.Hour
	defaultIssuer          = "thiam"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type JWTService struct {
	config JWTConfig
}

func NewJWTService(config JWTConfig) (*JWTService, error) {
	if config.Secret == "" {
		return nil, ErrMissingSecret
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

func (s *JWTService) GetRefreshTokenDuration() time.Duration {
	return s.config.RefreshTokenDuration
}
