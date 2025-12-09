package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		app.Use(RequestID())
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString(GetRequestID(c))
		})

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		resp, err := app.Test(req)

		require.NoError(t, err)

		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)

		requestID := string(body)

		_, err = uuid.Parse(requestID)
		assert.NoError(t, err, "should be valid UUID")

		assert.Equal(t, requestID, resp.Header.Get(RequestIDHeader))
	})

	t.Run("uses existing request ID from header", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		app.Use(RequestID())
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString(GetRequestID(c))
		})

		existingID := "existing-request-id-123"
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set(RequestIDHeader, existingID)

		resp, err := app.Test(req)

		require.NoError(t, err)

		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)

		assert.Equal(t, existingID, string(body))
		assert.Equal(t, existingID, resp.Header.Get(RequestIDHeader))
	})
}

func TestGetRequestID(t *testing.T) {
	t.Parallel()

	t.Run("returns request ID from context", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()

		var capturedID string

		app.Use(RequestID())
		app.Get("/", func(c *fiber.Ctx) error {
			capturedID = GetRequestID(c)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		resp, err := app.Test(req)

		require.NoError(t, err)

		defer resp.Body.Close()

		assert.NotEmpty(t, capturedID)
	})

	t.Run("returns empty string when no request ID", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()

		var capturedID string

		app.Get("/", func(c *fiber.Ctx) error {
			capturedID = GetRequestID(c)

			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		resp, err := app.Test(req)

		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Empty(t, capturedID)
	})
}
