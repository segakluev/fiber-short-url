package save

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	resp "go-short-url/internal/lib/api/response"
	"go-short-url/internal/lib/logger/sl"
	"go-short-url/internal/lib/random"
	"go-short-url/internal/storage"

	"log/slog"
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

func New(log *slog.Logger, urlSaver URLSaver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", c.Locals("requestid").(string)),
		)

		var req Request

		if err := c.BodyParser(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			return c.Status(fiber.StatusBadRequest).JSON(resp.Error("failed to decode request"))
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))
			return c.Status(fiber.StatusBadRequest).JSON(resp.ValidationError(err.(validator.ValidationErrors)))
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			return c.Status(fiber.StatusConflict).JSON(resp.Error("url already exists"))
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			return c.Status(fiber.StatusInternalServerError).JSON(resp.Error("failed to add url"))
		}

		log.Info("url added", slog.Int64("id", id))

		return c.JSON(Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
