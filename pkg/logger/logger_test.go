package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"unknown defaults to info", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := New(tt.level)

			require.NotNil(t, l)
			assert.NotNil(t, l.logger)
		})
	}
}

func TestLogger_WithField(t *testing.T) {
	t.Parallel()

	l := New("info")

	result := l.WithField("key", "value")

	require.NotNil(t, result)
	assert.NotEqual(t, l, result)
}

func TestLogger_WithFields(t *testing.T) {
	t.Parallel()

	l := New("info")

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	result := l.WithFields(fields)

	require.NotNil(t, result)
	assert.NotEqual(t, l, result)
}

func TestLogger_WithRequestID(t *testing.T) {
	t.Parallel()

	l := New("info")

	result := l.WithRequestID("test-request-id")

	require.NotNil(t, result)
	assert.NotEqual(t, l, result)
}

func TestLogger_WithContext(t *testing.T) {
	t.Parallel()

	l := New("info")

	t.Run("with request ID in context", func(t *testing.T) {
		t.Parallel()

		ctx := ContextWithRequestID(context.Background(), "ctx-request-id")

		result := l.WithContext(ctx)

		require.NotNil(t, result)
		assert.NotEqual(t, l, result)
	})

	t.Run("without request ID in context", func(t *testing.T) {
		t.Parallel()

		result := l.WithContext(context.Background())

		require.NotNil(t, result)
	})
}

func TestContextWithRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	requestID := "test-123"

	newCtx := ContextWithRequestID(ctx, requestID)

	assert.Equal(t, requestID, RequestIDFromContext(newCtx))
}

func TestRequestIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("with request ID", func(t *testing.T) {
		t.Parallel()

		ctx := ContextWithRequestID(context.Background(), "my-id")

		result := RequestIDFromContext(ctx)

		assert.Equal(t, "my-id", result)
	})

	t.Run("without request ID", func(t *testing.T) {
		t.Parallel()

		result := RequestIDFromContext(context.Background())

		assert.Empty(t, result)
	})
}
