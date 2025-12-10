package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/auth"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var errRepo = errors.New("repository error")

type mockUserRepo struct {
	createFunc      func(ctx context.Context, user *auth.User) error
	getByIDFunc     func(ctx context.Context, id uuid.UUID) (*auth.User, error)
	getByEmailFunc  func(ctx context.Context, email string) (*auth.User, error)
	updateFunc      func(ctx context.Context, user *auth.User) error
	existsByEmailFn func(ctx context.Context, email string) (bool, error)
	createdUser     *auth.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *auth.User) error {
	m.createdUser = user

	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}

	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}

	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *auth.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}

	return nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(ctx, email)
	}

	return false, nil
}

type mockRefreshTokenRepo struct {
	createFunc         func(ctx context.Context, token *auth.RefreshToken) error
	getByTokenHashFunc func(ctx context.Context, tokenHash string) (*auth.RefreshToken, error)
	revokeFunc         func(ctx context.Context, id uuid.UUID) error
	revokeByFamilyIDFn func(ctx context.Context, familyID uuid.UUID) error
	createdToken       *auth.RefreshToken
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *auth.RefreshToken) error {
	m.createdToken = token

	if m.createFunc != nil {
		return m.createFunc(ctx, token)
	}

	token.ID = uuid.New()
	token.FamilyID = uuid.New()
	token.CreatedAt = time.Now()

	return nil
}

func (m *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	if m.getByTokenHashFunc != nil {
		return m.getByTokenHashFunc(ctx, tokenHash)
	}

	return nil, nil
}

func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, id)
	}

	return nil
}

func (m *mockRefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID uuid.UUID) error {
	if m.revokeByFamilyIDFn != nil {
		return m.revokeByFamilyIDFn(ctx, familyID)
	}

	return nil
}

func newTestJWTService(t *testing.T) *pkgauth.JWTService {
	t.Helper()

	svc, err := pkgauth.NewJWTService(pkgauth.JWTConfig{
		Secret: "test-secret-key-for-testing-only-32-chars",
	})
	require.NoError(t, err)

	return svc
}

//nolint:funlen // table-driven tests are verbose by design
func TestRegisterUseCase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            authuc.RegisterInput
		userRepo         *mockUserRepo
		refreshTokenRepo *mockRefreshTokenRepo
		wantErr          error
		checkOutput      func(t *testing.T, output *authuc.RegisterOutput)
	}{
		{
			name: "success with email and password",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          nil,
			checkOutput: func(t *testing.T, output *authuc.RegisterOutput) {
				t.Helper()
				require.NotNil(t, output)
				require.NotNil(t, output.User)
				require.Equal(t, "test@example.com", output.User.Email)
				require.NotNil(t, output.User.PasswordHash)
				require.Equal(t, auth.UserStatusPendingVerification, output.User.Status)
				require.Nil(t, output.User.Name)
				require.NotNil(t, output.TokenPair)
				require.NotEmpty(t, output.TokenPair.AccessToken)
				require.NotEmpty(t, output.TokenPair.RefreshToken)
				require.Equal(t, "Bearer", output.TokenPair.TokenType)
			},
		},
		{
			name: "success with name",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "John Doe",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          nil,
			checkOutput: func(t *testing.T, output *authuc.RegisterOutput) {
				t.Helper()
				require.NotNil(t, output)
				require.NotNil(t, output.User)
				require.NotNil(t, output.User.Name)
				require.Equal(t, "John Doe", *output.User.Name)
			},
		},
		{
			name: "email normalization - uppercase",
			input: authuc.RegisterInput{
				Email:    "TEST@EXAMPLE.COM",
				Password: "password123",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          nil,
			checkOutput: func(t *testing.T, output *authuc.RegisterOutput) {
				t.Helper()
				require.Equal(t, "test@example.com", output.User.Email)
			},
		},
		{
			name: "email normalization - trim spaces",
			input: authuc.RegisterInput{
				Email:    "  test@example.com  ",
				Password: "password123",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          nil,
			checkOutput: func(t *testing.T, output *authuc.RegisterOutput) {
				t.Helper()
				require.Equal(t, "test@example.com", output.User.Email)
			},
		},
		{
			name: "invalid email - missing @",
			input: authuc.RegisterInput{
				Email:    "invalidemail",
				Password: "password123",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          authuc.ErrInvalidEmail,
		},
		{
			name: "password too short",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "short",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          authuc.ErrPasswordTooShort,
		},
		{
			name: "password exactly 8 characters - valid",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "12345678",
			},
			userRepo:         &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          nil,
			checkOutput: func(t *testing.T, output *authuc.RegisterOutput) {
				t.Helper()
				require.NotNil(t, output)
			},
		},
		{
			name: "email already exists",
			input: authuc.RegisterInput{
				Email:    "existing@example.com",
				Password: "password123",
			},
			userRepo: &mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ string) (bool, error) {
					return true, nil
				},
			},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          authuc.ErrEmailAlreadyExists,
		},
		{
			name: "error checking email exists",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			userRepo: &mockUserRepo{
				existsByEmailFn: func(_ context.Context, _ string) (bool, error) {
					return false, errRepo
				},
			},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          errRepo,
		},
		{
			name: "error creating user",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			userRepo: &mockUserRepo{
				createFunc: func(_ context.Context, _ *auth.User) error {
					return errRepo
				},
			},
			refreshTokenRepo: &mockRefreshTokenRepo{},
			wantErr:          errRepo,
		},
		{
			name: "error creating refresh token",
			input: authuc.RegisterInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			userRepo: &mockUserRepo{},
			refreshTokenRepo: &mockRefreshTokenRepo{
				createFunc: func(_ context.Context, _ *auth.RefreshToken) error {
					return errRepo
				},
			},
			wantErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jwtService := newTestJWTService(t)
			uc := authuc.NewRegisterUseCase(tt.userRepo, tt.refreshTokenRepo, jwtService)

			output, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, output)

				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestRegisterUseCase_PasswordHashing(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{}
	refreshTokenRepo := &mockRefreshTokenRepo{}
	jwtService := newTestJWTService(t)

	uc := authuc.NewRegisterUseCase(userRepo, refreshTokenRepo, jwtService)

	input := authuc.RegisterInput{
		Email:    "test@example.com",
		Password: "mySecurePassword123",
	}

	_, err := uc.Execute(context.Background(), input)
	require.NoError(t, err)

	require.NotNil(t, userRepo.createdUser)
	require.NotNil(t, userRepo.createdUser.PasswordHash)
	require.NotEqual(t, input.Password, *userRepo.createdUser.PasswordHash)

	valid, err := pkgauth.VerifyPassword(input.Password, *userRepo.createdUser.PasswordHash)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestRegisterUseCase_RefreshTokenCreation(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{}
	refreshTokenRepo := &mockRefreshTokenRepo{}
	jwtService := newTestJWTService(t)

	uc := authuc.NewRegisterUseCase(userRepo, refreshTokenRepo, jwtService)

	input := authuc.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := uc.Execute(context.Background(), input)
	require.NoError(t, err)

	require.NotNil(t, refreshTokenRepo.createdToken)
	require.Equal(t, userRepo.createdUser.ID, refreshTokenRepo.createdToken.UserID)
	require.Equal(t, 1, refreshTokenRepo.createdToken.Generation)
	require.NotEmpty(t, refreshTokenRepo.createdToken.TokenHash)

	expectedHash := pkgauth.HashToken(output.TokenPair.RefreshToken)
	require.Equal(t, expectedHash, refreshTokenRepo.createdToken.TokenHash)
}

func TestRegisterUseCase_TokenValidation(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{}
	refreshTokenRepo := &mockRefreshTokenRepo{}
	jwtService := newTestJWTService(t)

	uc := authuc.NewRegisterUseCase(userRepo, refreshTokenRepo, jwtService)

	input := authuc.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	output, err := uc.Execute(context.Background(), input)
	require.NoError(t, err)

	accessClaims, err := jwtService.ValidateAccessToken(output.TokenPair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, userRepo.createdUser.ID, accessClaims.UserID)
	require.Equal(t, "test@example.com", accessClaims.Email)

	refreshClaims, err := jwtService.ValidateRefreshToken(output.TokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, userRepo.createdUser.ID, refreshClaims.UserID)
	require.Equal(t, "test@example.com", refreshClaims.Email)
}
