package apperror_test

import (
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errDatabaseConnection = errors.New("database connection failed")
	errRandom             = errors.New("random error")
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *apperror.Error
		expected string
	}{
		{
			name:     "with operation",
			err:      apperror.NotFound("user not found", apperror.WithOp("UserRepo.GetByID")),
			expected: "UserRepo.GetByID: user not found",
		},
		{
			name:     "without operation",
			err:      apperror.NotFound("user not found"),
			expected: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	err := apperror.Internal("failed to fetch user", apperror.WithCause(errDatabaseConnection))

	assert.ErrorIs(t, err, errDatabaseConnection)
}

func TestError_Kind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *apperror.Error
		expected apperror.Kind
	}{
		{"validation error", apperror.Validation("invalid email"), apperror.KindValidation},
		{"not found error", apperror.NotFound("user not found"), apperror.KindNotFound},
		{"conflict error", apperror.Conflict("email already exists"), apperror.KindConflict},
		{"unauthorized error", apperror.Unauthorized("invalid credentials"), apperror.KindUnauthorized},
		{"forbidden error", apperror.Forbidden("access denied"), apperror.KindForbidden},
		{"internal error", apperror.Internal("unexpected error"), apperror.KindInternal},
		{"external error", apperror.External("payment gateway unavailable"), apperror.KindExternal},
		{"timeout error", apperror.Timeout("request timed out"), apperror.KindTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.err.Kind())
		})
	}
}

func TestError_Code(t *testing.T) {
	t.Parallel()

	t.Run("default code from kind", func(t *testing.T) {
		t.Parallel()

		err := apperror.NotFound("user not found")
		assert.Equal(t, "NOT_FOUND", err.Code())
	})

	t.Run("custom code", func(t *testing.T) {
		t.Parallel()

		err := apperror.NotFound("user not found", apperror.WithCode("USER_NOT_FOUND"))
		assert.Equal(t, "USER_NOT_FOUND", err.Code())
	})
}

func TestError_Fields(t *testing.T) {
	t.Parallel()

	t.Run("single field", func(t *testing.T) {
		t.Parallel()

		err := apperror.Validation("invalid email", apperror.WithField("email", "must be valid email"))

		fields := err.Fields()
		require.NotNil(t, fields)
		assert.Equal(t, "must be valid email", fields["email"])
	})

	t.Run("multiple fields", func(t *testing.T) {
		t.Parallel()

		err := apperror.Validation("validation failed", apperror.WithFields(map[string]string{
			"email": "must be valid email",
			"name":  "is required",
		}))

		fields := err.Fields()
		require.NotNil(t, fields)
		assert.Equal(t, "must be valid email", fields["email"])
		assert.Equal(t, "is required", fields["name"])
	})
}

func TestGetKind(t *testing.T) {
	t.Parallel()

	t.Run("app error", func(t *testing.T) {
		t.Parallel()

		err := apperror.NotFound("not found")
		assert.Equal(t, apperror.KindNotFound, apperror.GetKind(err))
	})

	t.Run("non-app error", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, apperror.KindUnknown, apperror.GetKind(errRandom))
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, apperror.KindUnknown, apperror.GetKind(nil))
	})
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsNotFound(apperror.NotFound("not found")))
	assert.False(t, apperror.IsNotFound(apperror.Validation("invalid")))
}

func TestIsValidation(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsValidation(apperror.Validation("invalid")))
	assert.False(t, apperror.IsValidation(apperror.NotFound("not found")))
}

func TestIsConflict(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsConflict(apperror.Conflict("conflict")))
}

func TestIsUnauthorized(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsUnauthorized(apperror.Unauthorized("unauthorized")))
}

func TestIsForbidden(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsForbidden(apperror.Forbidden("forbidden")))
}

func TestIsInternal(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsInternal(apperror.Internal("internal")))
}

func TestIsExternal(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsExternal(apperror.External("external")))
}

func TestIsTimeout(t *testing.T) {
	t.Parallel()

	assert.True(t, apperror.IsTimeout(apperror.Timeout("timeout")))
}

func TestAsAppError(t *testing.T) {
	t.Parallel()

	t.Run("is app error", func(t *testing.T) {
		t.Parallel()

		err := apperror.NotFound("not found")

		appErr, ok := apperror.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperror.KindNotFound, appErr.Kind())
	})

	t.Run("not app error", func(t *testing.T) {
		t.Parallel()

		appErr, ok := apperror.AsAppError(errRandom)
		assert.False(t, ok)
		assert.Nil(t, appErr)
	})
}

func TestKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind     apperror.Kind
		expected string
	}{
		{apperror.KindValidation, "VALIDATION_ERROR"},
		{apperror.KindNotFound, "NOT_FOUND"},
		{apperror.KindConflict, "CONFLICT"},
		{apperror.KindUnauthorized, "UNAUTHORIZED"},
		{apperror.KindForbidden, "FORBIDDEN"},
		{apperror.KindInternal, "INTERNAL_ERROR"},
		{apperror.KindExternal, "EXTERNAL_SERVICE_ERROR"},
		{apperror.KindTimeout, "TIMEOUT"},
		{apperror.KindUnknown, "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.kind.String())
		})
	}
}
