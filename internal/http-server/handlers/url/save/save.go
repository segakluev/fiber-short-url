package save

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/lib/random"
	"go-short-url/internal/storage"
	"go-short-url/internal/storage/cache"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver, c *cache.MemoryCache) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const op = "handlers.url.save.New"

		reqID, _ := ctx.Locals("requestid").(string)
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqID),
		)

		var req Request
		if err := ctx.BodyParser(&req); err != nil {
			logger.Error("failed to decode request body", sl.Err(err))
			return ctx.Status(fiber.StatusBadRequest).JSON(resp.Error("failed to decode request"))
		}

		if err := validator.New().Struct(req); err != nil {
			logger.Info("invalid request", sl.Err(err))
			return ctx.Status(fiber.StatusBadRequest).JSON(resp.ValidationError(err.(validator.ValidationErrors)))
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			logger.Info("url already exists", slog.String("url", req.URL))
			return ctx.Status(fiber.StatusConflict).JSON(resp.Error("url already exists"))
		}
		if err != nil {
			logger.Error("failed to add url", sl.Err(err))
			return ctx.Status(fiber.StatusInternalServerError).JSON(resp.Error("failed to add url"))
		}

		// warm up cache
		if c != nil {
			c.Set(alias, req.URL)
		}

		logger.Info("url added", slog.Int64("id", id), slog.String("alias", alias))
		return ctx.JSON(Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
