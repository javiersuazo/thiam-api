package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/evrone/go-clean-template/internal/controller/http/v1"
	"github.com/evrone/go-clean-template/internal/entity/auth"
	authuc "github.com/evrone/go-clean-template/internal/usecase/auth"
	pkgauth "github.com/evrone/go-clean-template/pkg/auth"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDatabase = errors.New("database error")

type testUserRepo struct {
	exists    bool
	existsErr error
	createErr error
}

func (r *testUserRepo) Create(_ context.Context, user *auth.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return nil
}

func (r *testUserRepo) CreateTx(_ context.Context, _ postgres.DBTX, user *auth.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return nil
}

func (r *testUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*auth.User, error) {
	return nil, nil
}

func (r *testUserRepo) GetByEmail(_ context.Context, _ string) (*auth.User, error) {
	return nil, nil
}

func (r *testUserRepo) Update(_ context.Context, _ *auth.User) error {
	return nil
}

func (r *testUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error) {
	if r.existsErr != nil {
		return false, r.existsErr
	}

	return r.exists, nil
}

type testRefreshTokenRepo struct{}

func (r *testRefreshTokenRepo) Create(_ context.Context, token *auth.RefreshToken) error {
	token.ID = uuid.New()
	token.FamilyID = uuid.New()
	token.CreatedAt = time.Now()

	return nil
}

func (r *testRefreshTokenRepo) CreateTx(_ context.Context, _ postgres.DBTX, token *auth.RefreshToken) error {
	token.ID = uuid.New()
	token.FamilyID = uuid.New()
	token.CreatedAt = time.Now()

	return nil
}

func (r *testRefreshTokenRepo) GetByTokenHash(_ context.Context, _ string) (*auth.RefreshToken, error) {
	return nil, nil
}

func (r *testRefreshTokenRepo) Revoke(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *testRefreshTokenRepo) RevokeByFamilyID(_ context.Context, _ uuid.UUID) error {
	return nil
}

type testTxManager struct{}

func (m *testTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func newTestApp(t *testing.T, userRepo *testUserRepo) *fiber.App {
	t.Helper()

	jwtService, err := pkgauth.NewJWTService(pkgauth.JWTConfig{
		Secret: "test-secret-key-for-testing-only-32-chars",
	})
	require.NoError(t, err)

	txManager := &testTxManager{}
	uc := authuc.NewRegisterUseCase(userRepo, &testRefreshTokenRepo{}, jwtService, txManager)

	app := fiber.New()
	group := app.Group("/v1")
	v1.NewAuthRoutes(group, uc)

	return app
}

func checkResponseBody(t *testing.T, resp *http.Response, check func(response map[string]interface{})) {
	t.Helper()

	respBody, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)

	var response map[string]interface{}

	err := json.Unmarshal(respBody, &response)
	require.NoError(t, err)

	check(response)
}

//nolint:funlen // table-driven tests are verbose by design
func TestAuthRoutes_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		userRepo   *testUserRepo
		wantStatus int
		check      func(t *testing.T, response map[string]interface{})
	}{
		{
			name:       "success with name",
			body:       `{"email":"test@example.com","password":"Password123!","name":"John Doe"}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, response map[string]interface{}) {
				t.Helper()
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["refresh_token"])
				assert.Equal(t, "Bearer", response["token_type"])

				user, ok := response["user"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test@example.com", user["email"])
				assert.Equal(t, "John Doe", user["name"])
			},
		},
		{
			name:       "success without name",
			body:       `{"email":"test@example.com","password":"Password123!"}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, response map[string]interface{}) {
				t.Helper()
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["refresh_token"])
				assert.Equal(t, "Bearer", response["token_type"])

				user, ok := response["user"].(map[string]interface{})
				require.True(t, ok)
				assert.Nil(t, user["name"])
			},
		},
		{
			name:       "invalid request body",
			body:       `{invalid json}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "email already exists",
			body:       `{"email":"existing@example.com","password":"Password123!"}`,
			userRepo:   &testUserRepo{exists: true},
			wantStatus: http.StatusConflict,
			check: func(t *testing.T, response map[string]interface{}) {
				t.Helper()
				assert.Equal(t, "EMAIL_ALREADY_EXISTS", response["code"])
			},
		},
		{
			name:       "invalid email format",
			body:       `{"email":"invalidemail","password":"Password123!"}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "password too short",
			body:       `{"email":"test@example.com","password":"Sh0rt!"}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "password too weak",
			body:       `{"email":"test@example.com","password":"password123"}`,
			userRepo:   &testUserRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "internal server error",
			body:       `{"email":"test@example.com","password":"Password123!"}`,
			userRepo:   &testUserRepo{existsErr: errDatabase},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := newTestApp(t, tt.userRepo)

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.check != nil {
				checkResponseBody(t, resp, func(response map[string]interface{}) {
					tt.check(t, response)
				})
			}
		})
	}
}
