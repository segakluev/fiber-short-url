package logger

import (
	"fmt"
	"time"

	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func New(log *slog.Logger) fiber.Handler {
	base := log.With(slog.String("component", "middleware/logger"))

	base.Info("logger middleware enabled")

	return func(c *fiber.Ctx) error {
		entry := base.With(
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("remote_addr", c.IP()),
			slog.String("user_agent", c.Get("User-Agent")),
			slog.String("request_id", fmt.Sprint(c.Locals("requestid"))),
		)

		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		bytes := c.Response().Header.ContentLength()

		entry.Info("request completed",
			slog.Int("status", status),
			slog.Int("bytes", bytes),
			slog.String("duration", time.Since(start).String()),
		)

		return err
	}
}
