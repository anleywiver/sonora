package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"sonora.dev/go-core/config"
	"sonora.dev/go-core/infrastructure/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres pool: %v", err)
	}
	defer pool.Close()

	gormDB, err := postgres.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("gorm: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("gorm underlying db: %v", err)
	}
	defer sqlDB.Close()

	app := fiber.New(fiber.Config{
		AppName: "Sonora API v1",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:3001",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Health check - dipakai Docker Compose healthcheck & CI smoke test
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "sonora-api",
		})
	})

	api := app.Group("/api/v1")
	_ = api // route group siap diisi di Task berikutnya (auth, songs, dst)

	log.Fatal(app.Listen(":8080"))
}
