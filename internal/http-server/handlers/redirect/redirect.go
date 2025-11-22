package redirect

import (
	"errors"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/storage"

	"github.com/gofiber/fiber/v2"

	"log/slog"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const op = "handlers.url.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", c.Locals("requestid").(string)),
		)

		// Получаем alias из маршрута: /:alias
		alias := c.Params("alias")
		if alias == "" {
			log.Info("alias is empty")
			return c.Status(fiber.StatusBadRequest).JSON(resp.Error("invalid request"))
		}

		// Запрашиваем URL из БД
		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			return c.Status(fiber.StatusNotFound).JSON(resp.Error("not found"))
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			return c.Status(fiber.StatusInternalServerError).JSON(resp.Error("internal error"))
		}

		log.Info("got url", slog.String("url", resURL))

		// Делаем 302 редирект
		return c.Redirect(resURL, fiber.StatusFound)
	}
}
