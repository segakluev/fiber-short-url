package delete

import (
	"errors"
	"log/slog"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/storage"
	"go-short-url/internal/storage/cache"

	"github.com/gofiber/fiber/v2"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleter URLDeleter, c *cache.MemoryCache) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const op = "handlers.url.delete.New"

		reqID, _ := ctx.Locals("requestid").(string)
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqID),
		)

		alias := ctx.Params("alias")
		if alias == "" {
			logger.Info("empty alias")
			return ctx.Status(fiber.StatusBadRequest).JSON(resp.Error("invalid alias"))
		}

		err := deleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			logger.Info("alias not found", slog.String("alias", alias))
			return ctx.Status(fiber.StatusNotFound).JSON(resp.Error("alias not found"))
		}
		if err != nil {
			logger.Error("failed to delete alias", sl.Err(err))
			return ctx.Status(fiber.StatusInternalServerError).JSON(resp.Error("internal error"))
		}

		// remove from cache too
		if c != nil {
			c.Delete(alias)
		}

		logger.Info("alias deleted", slog.String("alias", alias))
		return ctx.JSON(resp.OK())
	}
}
