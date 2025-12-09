package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the HTTP header name for request ID.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the key used to store request ID in fiber.Ctx.Locals.
	RequestIDKey = "request_id"
)

// RequestID is a middleware that generates or extracts a request ID for each request.
// If the X-Request-ID header is present, it uses that value; otherwise generates a new UUID.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Locals(RequestIDKey, requestID)
		c.Set(RequestIDHeader, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from the fiber context.
// Returns an empty string if no request ID is set.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(RequestIDKey).(string); ok {
		return id
	}

	return ""
}
