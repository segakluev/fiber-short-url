package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"go-short-url/internal/config"
	"go-short-url/internal/storage/sqlite"

	mwLogger "go-short-url/internal/http-server/middleware/logger"

	redirectHandler "go-short-url/internal/http-server/handlers/redirect"
	deleteHandler "go-short-url/internal/http-server/handlers/url/delete"
	saveHandler "go-short-url/internal/http-server/handlers/url/save"
)

func main() {
	// === 1. Load config ===
	cfg := config.MustLoad()

	// === 2. Create logger (slog) ===
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// === 3. Init SQLite Storage ===
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", slog.Any("error", err))
		os.Exit(1)
	}
	defer storage.Close()

	// === 4. Init Fiber ===
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// === 5. Middleware ===
	app.Use(requestid.New())   // request_id → доступен в c.Locals("requestid")
	app.Use(recover.New())     // ловит паники
	app.Use(mwLogger.New(log)) // кастомный логгер

	// === 6. Routes ===

	// POST /url — создать сокращённую ссылку
	app.Post("/url", saveHandler.New(log, storage))

	// GET /:alias — редирект
	app.Get("/:alias", redirectHandler.New(log, storage))

	// DELETE /delete/:alias — удалить alias
	app.Delete("/delete/:alias", deleteHandler.New(log, storage))

	// === 7. Start server ===
	log.Info("server starting", slog.String("address", cfg.HTTPServer.Address))

	if err := app.Listen(cfg.HTTPServer.Address); err != nil {
		log.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
