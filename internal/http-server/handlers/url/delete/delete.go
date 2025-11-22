package delete

import (
	"errors"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/storage"

	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", c.Locals("requestid").(string)),
		)

		alias := c.Params("alias")
		if alias == "" {
			log.Warn("alias is empty")
			return c.Status(fiber.StatusBadRequest).JSON(resp.Error("alias is required"))
		}

		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Warn("alias not found", slog.String("alias", alias))
			return c.Status(fiber.StatusNotFound).JSON(resp.Error("alias not found"))
		}
		if err != nil {
			log.Error("failed to delete", sl.Err(err))
			return c.Status(fiber.StatusInternalServerError).JSON(resp.Error("failed to delete"))
		}

		log.Info("alias deleted", slog.String("alias", alias))

		return c.JSON(resp.OK())
	}
}
