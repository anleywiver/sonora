package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	appauth "sonora.dev/go-core/application/auth"
	"sonora.dev/go-core/config"
	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/jwt"
	"sonora.dev/go-core/infrastructure/oauth"
	"sonora.dev/go-core/infrastructure/postgres"
	"sonora.dev/go-core/infrastructure/postgres/repository"

	"sonora.dev/backend/internal/http/handlers"
	"sonora.dev/backend/internal/http/middleware"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
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

	userRepo := repository.NewUserRepository(gormDB)
	deviceRepo := repository.NewDeviceRepository(gormDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(gormDB)

	googleClient := oauth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	jwtIssuer := jwt.NewIssuer(cfg.JWTAccessSecret, accessTokenTTL)

	authService := appauth.NewService(userRepo, deviceRepo, refreshTokenRepo, googleClient, jwtIssuer, refreshTokenTTL)
	authHandler := handlers.NewAuthHandler(authService, cfg.FrontendURL)
	deviceHandler := handlers.NewDeviceHandler(authService)

	requireAuth := middleware.RequireAuth(jwtIssuer)
	requireOwner := middleware.RequireRole(string(identity.RoleOwner))

	app := fiber.New(fiber.Config{
		AppName: "Sonora API v1",
	})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Health check - dipakai Docker Compose healthcheck & CI smoke test
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "sonora-api",
		})
	})

	api := app.Group("/api/v1")

	authGroup := api.Group("/auth")
	authGroup.Get("/google", authHandler.GoogleLogin)
	authGroup.Get("/google/callback", authHandler.GoogleCallback)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", requireAuth, authHandler.Logout)
	authGroup.Post("/logout-all", requireAuth, requireOwner, authHandler.LogoutAll)
	authGroup.Get("/me", requireAuth, authHandler.Me)

	api.Get("/devices", requireAuth, deviceHandler.List)
	api.Delete("/devices/:id", requireAuth, deviceHandler.Delete)

	log.Fatal(app.Listen(":8080"))
}
