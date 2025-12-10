package v1

import (
	"errors"
	"net/http"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	"github.com/evrone/go-clean-template/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type authRoutes struct {
	registerUC *authuc.RegisterUseCase
}

func NewAuthRoutes(group fiber.Router, registerUC *authuc.RegisterUseCase) {
	r := &authRoutes{
		registerUC: registerUC,
	}

	authGroup := group.Group("/auth")
	authGroup.Post("/register", r.register)
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          *string `json:"name,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	PhoneNumber   *string `json:"phone_number,omitempty"`
	PhoneVerified bool    `json:"phone_verified"`
	MFAEnabled    bool    `json:"mfa_enabled"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
	TokenType    string       `json:"token_type"`
}

func (r *authRoutes) register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	input := authuc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	output, err := r.registerUC.Execute(c.Context(), input)
	if err != nil {
		return r.handleRegisterError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(AuthResponse{
		User:         toUserResponse(output.User),
		AccessToken:  output.TokenPair.AccessToken,
		RefreshToken: output.TokenPair.RefreshToken,
		ExpiresIn:    output.TokenPair.ExpiresIn,
		TokenType:    output.TokenPair.TokenType,
	})
}

func (r *authRoutes) handleRegisterError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authuc.ErrEmailAlreadyExists):
		return ErrorResponse(c, apperror.Conflict("Email is already registered", apperror.WithCode("EMAIL_ALREADY_EXISTS")))
	case errors.Is(err, authuc.ErrInvalidEmail):
		return ErrorResponse(c, apperror.Validation("Invalid email format"))
	case errors.Is(err, authuc.ErrPasswordTooShort):
		return ErrorResponse(c, apperror.Validation("Password must be at least 8 characters"))
	case errors.Is(err, authuc.ErrPasswordTooWeak):
		return ErrorResponse(c, apperror.Validation("Password must contain uppercase, lowercase, digit, and special character"))
	default:
		return ErrorResponse(c, apperror.Internal("An unexpected error occurred"))
	}
}

func toUserResponse(u *auth.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
		PhoneNumber:   u.PhoneNumber,
		PhoneVerified: u.PhoneVerified,
		MFAEnabled:    false, // Computed from MFA tables - will implement later
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
