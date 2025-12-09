package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen // table-driven test
func TestRedactFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "redacts password field",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "john",
				"password": "[REDACTED]",
			},
		},
		{
			name: "redacts token field",
			input: map[string]interface{}{
				"token": "abc123",
			},
			expected: map[string]interface{}{
				"token": "[REDACTED]",
			},
		},
		{
			name: "redacts api_key field",
			input: map[string]interface{}{
				"api_key": "key-123",
			},
			expected: map[string]interface{}{
				"api_key": "[REDACTED]",
			},
		},
		{
			name: "masks email in string value",
			input: map[string]interface{}{
				"message": "User john@example.com logged in",
			},
			expected: map[string]interface{}{
				"message": "User j***n@example.com logged in",
			},
		},
		{
			name: "redacts credit card numbers",
			input: map[string]interface{}{
				"payment": "Card 4111111111111111 used",
			},
			expected: map[string]interface{}{
				"payment": "Card [REDACTED] used",
			},
		},
		{
			name: "preserves non-sensitive fields",
			input: map[string]interface{}{
				"status": 200,
				"method": "GET",
			},
			expected: map[string]interface{}{
				"status": 200,
				"method": "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := RedactFields(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "masks normal email",
			email:    "john@example.com",
			expected: "j***n@example.com",
		},
		{
			name:     "masks short local part",
			email:    "ab@example.com",
			expected: "**@example.com",
		},
		{
			name:     "masks single char local",
			email:    "a@example.com",
			expected: "**@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := maskEmail(tt.email)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	t.Parallel()

	sensitiveKeys := []string{
		"password", "PASSWORD", "Password",
		"secret", "token", "authorization",
		"api_key", "apikey", "credit_card", "ssn", "cvv",
	}

	for _, key := range sensitiveKeys {
		assert.True(t, isSensitiveKey(key), "expected %s to be sensitive", key)
	}

	nonSensitiveKeys := []string{"username", "email", "status", "method"}

	for _, key := range nonSensitiveKeys {
		assert.False(t, isSensitiveKey(key), "expected %s to not be sensitive", key)
	}
}
