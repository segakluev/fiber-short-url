package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"go-short-url/internal/config"
	"go-short-url/internal/storage/cache"
	"go-short-url/internal/storage/sqlite"

	mwLogger "go-short-url/internal/http-server/middleware/logger"

	redirectHandler "go-short-url/internal/http-server/handlers/redirect"
	deleteHandler "go-short-url/internal/http-server/handlers/url/delete"
	saveHandler "go-short-url/internal/http-server/handlers/url/save"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// === cache ===
	cacheTTL := 10 * time.Minute
	cacheCleanup := 1 * time.Minute
	memCache := cache.New(cacheTTL)
	memCache.StartEvictionWorker(cacheCleanup)
	defer memCache.Stop()

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(mwLogger.New(log))

	// inject cache into handlers
	app.Post("/url", saveHandler.New(log, db, memCache))
	app.Get("/:alias", redirectHandler.New(log, db, memCache))
	app.Delete("/delete/:alias", deleteHandler.New(log, db, memCache))

	log.Info("server starting", slog.String("address", cfg.HTTPServer.Address))
	if err := app.Listen(cfg.HTTPServer.Address); err != nil {
		log.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
