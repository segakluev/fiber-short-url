package redirect

import (
	"errors"
	"log/slog"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/storage"
	"go-short-url/internal/storage/cache"

	"github.com/gofiber/fiber/v2"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, db URLGetter, c *cache.MemoryCache) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const op = "handlers.url.redirect"

		reqID, _ := ctx.Locals("requestid").(string)
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqID),
		)

		alias := ctx.Params("alias")
		if alias == "" {
			logger.Warn("empty alias")
			return ctx.Status(fiber.StatusBadRequest).JSON(resp.Error("invalid alias"))
		}

		// 1) Попробовать из кеша
		if c != nil {
			if url, ok := c.Get(alias); ok {
				logger.Debug("cache hit", slog.String("alias", alias))
				return ctx.Redirect(url, fiber.StatusFound)
			}
			logger.Debug("cache miss", slog.String("alias", alias))
		}

		// 2) Подгрузить из БД
		target, err := db.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			logger.Info("alias not found", slog.String("alias", alias))
			return ctx.Status(fiber.StatusNotFound).JSON(resp.Error("not found"))
		}
		if err != nil {
			logger.Error("failed to get url from db", sl.Err(err))
			return ctx.Status(fiber.StatusInternalServerError).JSON(resp.Error("internal error"))
		}

		// 3) Кладём в кеш (если он есть) и редиректим
		if c != nil {
			c.Set(alias, target)
		}

		logger.Info("redirecting", slog.String("alias", alias), slog.String("to", target))
		return ctx.Redirect(target, fiber.StatusFound)
	}
}
