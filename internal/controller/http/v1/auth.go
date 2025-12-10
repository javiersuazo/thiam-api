package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	"github.com/evrone/go-clean-template/internal/usecase"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	"github.com/evrone/go-clean-template/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

const bearerTokenParts = 2

type authRoutes struct {
	authUC usecase.AuthUseCase
}

func NewAuthRoutes(group fiber.Router, authUC usecase.AuthUseCase) {
	r := &authRoutes{authUC: authUC}

	authGroup := group.Group("/auth")
	authGroup.Post("/register", r.register)
	authGroup.Post("/login", r.login)
	authGroup.Post("/logout", r.logout)
	authGroup.Post("/refresh", r.refresh)
	authGroup.Get("/me", r.me)
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
	TokenType    string       `json:"token_type"`
}

type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          *string `json:"name,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	Status        string  `json:"status"`
}

func (r *authRoutes) register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	if req.Email == "" {
		return ValidationError(c, "Email is required")
	}

	if req.Password == "" {
		return ValidationError(c, "Password is required")
	}

	input := &authuc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	output, err := r.authUC.Register(c.Context(), input)
	if err != nil {
		return r.handleAuthError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(AuthResponse{
		User:         toUserResponse(output.User),
		AccessToken:  output.TokenPair.AccessToken,
		RefreshToken: output.TokenPair.RefreshToken,
		ExpiresAt:    output.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    output.TokenPair.TokenType,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *authRoutes) login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return ValidationError(c, "Email and password are required")
	}

	input := &authuc.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}

	output, err := r.authUC.Login(c.Context(), input)
	if err != nil {
		return r.handleAuthError(c, err)
	}

	return c.Status(http.StatusOK).JSON(AuthResponse{
		User:         toUserResponse(output.User),
		AccessToken:  output.TokenPair.AccessToken,
		RefreshToken: output.TokenPair.RefreshToken,
		ExpiresAt:    output.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    output.TokenPair.TokenType,
	})
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *authRoutes) logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	if req.RefreshToken == "" {
		return ValidationError(c, "Refresh token is required")
	}

	err := r.authUC.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		return r.handleAuthError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

func (r *authRoutes) refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	if req.RefreshToken == "" {
		return ValidationError(c, "Refresh token is required")
	}

	input := &authuc.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
		IPAddress:    c.IP(),
		UserAgent:    c.Get("User-Agent"),
	}

	output, err := r.authUC.RefreshToken(c.Context(), input)
	if err != nil {
		return r.handleAuthError(c, err)
	}

	return c.Status(http.StatusOK).JSON(RefreshResponse{
		AccessToken:  output.TokenPair.AccessToken,
		RefreshToken: output.TokenPair.RefreshToken,
		ExpiresAt:    output.TokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    output.TokenPair.TokenType,
	})
}

func (r *authRoutes) me(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ErrorResponse(c, apperror.Unauthorized("Authorization header is required"))
	}

	parts := strings.SplitN(authHeader, " ", bearerTokenParts)
	if len(parts) != bearerTokenParts || !strings.EqualFold(parts[0], "bearer") {
		return ErrorResponse(c, apperror.Unauthorized("Invalid authorization header format"))
	}

	claims, err := r.authUC.ValidateAccessToken(parts[1])
	if err != nil {
		return ErrorResponse(c, apperror.Unauthorized("Invalid or expired token"))
	}

	user, err := r.authUC.GetCurrentUser(c.Context(), claims.UserID)
	if err != nil {
		return r.handleAuthError(c, err)
	}

	return c.Status(http.StatusOK).JSON(UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
		Status:        string(user.Status),
	})
}

func (r *authRoutes) handleAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authuc.ErrEmailAlreadyExists):
		return ErrorResponse(c, apperror.Conflict("Email already registered"))
	case errors.Is(err, authuc.ErrInvalidCredentials):
		return ErrorResponse(c, apperror.Unauthorized("Invalid email or password"))
	case errors.Is(err, authuc.ErrAccountLocked):
		return ErrorResponse(c, apperror.Forbidden("Account is temporarily locked"))
	case errors.Is(err, authuc.ErrUserNotFound):
		return ErrorResponse(c, apperror.NotFound("User not found"))
	case errors.Is(err, authuc.ErrInvalidToken), errors.Is(err, authuc.ErrTokenRevoked):
		return ErrorResponse(c, apperror.Unauthorized("Invalid or expired token"))
	default:
		return ErrorResponse(c, apperror.Internal("An unexpected error occurred", apperror.WithCause(err)))
	}
}

func toUserResponse(u *auth.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
		Status:        string(u.Status),
	}
}
