package logger

import (
	"regexp"
	"strings"
)

//nolint:gochecknoglobals // package-level regex patterns and sensitive keys
var (
	sensitiveKeys = map[string]bool{
		"password":      true,
		"secret":        true,
		"token":         true,
		"authorization": true,
		"api_key":       true,
		"apikey":        true,
		"credit_card":   true,
		"ssn":           true,
		"cvv":           true,
	}

	emailRegex      = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	creditCardRegex = regexp.MustCompile(`\b\d{13,16}\b`)
)

const (
	redactedValue      = "[REDACTED]"
	expectedEmailParts = 2
	minLocalPartLength = 2
)

// RedactFields returns a copy of fields with sensitive data redacted.
// Sensitive keys (password, token, etc.) are replaced with "[REDACTED]".
// String values are scanned for emails and credit card numbers which are masked.
func RedactFields(fields map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))

	for k, v := range fields {
		if isSensitiveKey(k) {
			result[k] = redactedValue

			continue
		}

		if str, ok := v.(string); ok {
			result[k] = redactString(str)
		} else {
			result[k] = v
		}
	}

	return result
}

func isSensitiveKey(key string) bool {
	return sensitiveKeys[strings.ToLower(key)]
}

func redactString(s string) string {
	s = emailRegex.ReplaceAllStringFunc(s, maskEmail)
	s = creditCardRegex.ReplaceAllString(s, redactedValue)

	return s
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != expectedEmailParts {
		return email
	}

	local := parts[0]
	if len(local) <= minLocalPartLength {
		return "**@" + parts[1]
	}

	return local[:1] + "***" + local[len(local)-1:] + "@" + parts[1]
}

// WithRedactedFields returns a logger with the given fields after applying redaction.
func (l *Logger) WithRedactedFields(fields map[string]interface{}) Interface {
	return l.WithFields(RedactFields(fields))
}
