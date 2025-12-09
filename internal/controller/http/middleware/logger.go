package middleware

import (
	"net/http"
	"time"

	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func Logger(l logger.Interface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		requestID := GetRequestID(c)

		logFields := map[string]interface{}{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     status,
			"latency_ms": latency.Milliseconds(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
			"bytes_out":  len(c.Response().Body()),
		}

		reqLogger := l.WithRequestID(requestID).WithFields(logger.RedactFields(logFields))

		switch {
		case status >= http.StatusInternalServerError:
			reqLogger.Error("server error")
		case status >= http.StatusBadRequest:
			reqLogger.Warn("client error")
		default:
			reqLogger.Info("request completed")
		}

		return err
	}
}
