package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	appadminsongs "sonora.dev/go-core/application/adminsongs"
	appanalytics "sonora.dev/go-core/application/analytics"
	appsettings "sonora.dev/go-core/application/appsettings"
	appauth "sonora.dev/go-core/application/auth"
	appcatalog "sonora.dev/go-core/application/catalog"
	appdashboard "sonora.dev/go-core/application/dashboard"
	apphistory "sonora.dev/go-core/application/history"
	appingest "sonora.dev/go-core/application/ingest"
	appingestfilter "sonora.dev/go-core/application/ingestfilter"
	appingestsource "sonora.dev/go-core/application/ingestsource"
	applibrary "sonora.dev/go-core/application/library"
	applyrics "sonora.dev/go-core/application/lyrics"
	appplayback "sonora.dev/go-core/application/playback"
	appqueue "sonora.dev/go-core/application/queue"
	appsearch "sonora.dev/go-core/application/search"
	appstorage "sonora.dev/go-core/application/storageaccount"
	appusers "sonora.dev/go-core/application/users"
	"sonora.dev/go-core/config"
	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/idempotency"
	"sonora.dev/go-core/infrastructure/jwt"
	infralyrics "sonora.dev/go-core/infrastructure/lyrics"
	"sonora.dev/go-core/infrastructure/meilisearch"
	"sonora.dev/go-core/infrastructure/oauth"
	"sonora.dev/go-core/infrastructure/postgres"
	"sonora.dev/go-core/infrastructure/postgres/repository"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
	"sonora.dev/go-core/infrastructure/redis"
	"sonora.dev/go-core/infrastructure/wstoken"

	"sonora.dev/backend/internal/http/handlers"
	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/ws"
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
	appSettingsRepo := repository.NewAppSettingsRepository(gormDB)

	googleClient := oauth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	jwtIssuer := jwt.NewIssuer(cfg.JWTAccessSecret, accessTokenTTL)

	appSettingsService := appsettings.NewService(appSettingsRepo)
	adminSettingsHandler := handlers.NewAdminSettingsHandler(appSettingsService)

	authService := appauth.NewService(userRepo, deviceRepo, refreshTokenRepo, googleClient, jwtIssuer, refreshTokenTTL)
	authHandler := handlers.NewAuthHandler(authService, appSettingsService, cfg.FrontendURL, cfg.AdminURL)
	deviceHandler := handlers.NewDeviceHandler(authService)

	credentialsBox, err := crypto.NewBox(cfg.StorageCredentialsEncryptionKey)
	if err != nil {
		log.Fatalf("storage credentials box: %v", err)
	}
	queries := sqlc.New(pool)
	redisClient := redis.NewClient(cfg.RedisURL)
	idempotencyStore := idempotency.NewStore(redisClient)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	defer asynqClient.Close()

	meiliClient := meilisearch.NewClient(cfg.MeilisearchURL, cfg.MeilisearchAPIKey)

	ingestService := appingest.NewService(queries, credentialsBox, meiliClient, cfg.GoogleClientID, cfg.GoogleClientSecret)
	ingestHandler := handlers.NewIngestHandler(ingestService, asynqClient, idempotencyStore, cfg.IngestTmpDir)

	storageAccountService := appstorage.NewService(queries, credentialsBox, cfg.GoogleClientID, cfg.GoogleClientSecret)
	storageAccountHandler := handlers.NewStorageAccountHandler(storageAccountService)

	ingestSourceService := appingestsource.NewService(queries, credentialsBox, userRepo, ingestService, cfg.IngestTmpDir)
	ingestSourceHandler := handlers.NewIngestSourceHandler(ingestSourceService)

	analyticsService := appanalytics.NewService(queries)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	backupHandler := handlers.NewBackupHandler(asynqClient)

	dashboardService := appdashboard.NewService(queries)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	ingestFilterService := appingestfilter.NewService(queries)
	ingestFilterHandler := handlers.NewIngestFilterHandler(ingestFilterService)

	usersService := appusers.NewService(userRepo)
	usersHandler := handlers.NewUsersHandler(usersService)

	adminSongsService := appadminsongs.NewService(queries, meiliClient)
	adminSongsHandler := handlers.NewAdminSongsHandler(adminSongsService)

	catalogService := appcatalog.NewService(queries, credentialsBox, cfg.JWTAccessSecret, cfg.GoogleClientID, cfg.GoogleClientSecret)
	catalogHandler := handlers.NewCatalogHandler(catalogService)

	searchService := appsearch.NewService(meiliClient)
	searchHandler := handlers.NewSearchHandler(searchService, catalogService)

	playlistRepo := repository.NewPlaylistRepository(gormDB)
	favoriteRepo := repository.NewFavoriteRepository(gormDB)
	libraryService := applibrary.NewService(playlistRepo, favoriteRepo, catalogService)
	playlistHandler := handlers.NewPlaylistHandler(libraryService)
	favoriteHandler := handlers.NewFavoriteHandler(libraryService)

	historyService := apphistory.NewService(queries)
	historyHandler := handlers.NewHistoryHandler(historyService, catalogService)

	queueService := appqueue.NewService(queries, catalogService)
	queueHandler := handlers.NewQueueHandler(queueService)

	lrclibClient := infralyrics.NewLRCLIBClient()
	lyricsService := applyrics.NewService(queries, catalogService, lrclibClient)
	lyricsHandler := handlers.NewLyricsHandler(lyricsService)
	lyricsProviderHandler := handlers.NewLyricsProviderHandler(lyricsService)

	wsHub := ws.NewHub()
	wsTokens := wstoken.NewIssuer(redisClient)
	playbackService := appplayback.NewService(queries)
	wsHandler := handlers.NewWSHandler(wsTokens, wsHub, playbackService)
	playerHandler := handlers.NewPlayerHandler(playbackService, authService, wsHub)

	requireAuth := middleware.RequireAuth(jwtIssuer)
	requireOwner := middleware.RequireRole(string(identity.RoleOwner))
	maintenanceGate := middleware.MaintenanceGate(appSettingsService)

	app := fiber.New(fiber.Config{
		AppName: "Sonora API v1",
	})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(middleware.Metrics)
	// Sprint 12 (ADR 0006): origins come from config, not a hardcoded dev
	// string, so a real VPS deploy with real domains just works via .env.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL + "," + cfg.AdminURL,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	// Sprint 12 (ADR 0006): generous global ceiling — a safety net against
	// a looping/buggy client, not a real usage limit for personal scale.
	app.Use(limiter.New(limiter.Config{
		Max:        300,
		Expiration: time.Minute,
	}))

	// Health check - dipakai Docker Compose healthcheck & CI smoke test
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "sonora-api",
		})
	})

	// Sprint 13 (ADR 0007): no /api/v1 prefix, no auth — internal metrics
	// endpoint for a Prometheus scraper, not a public API route.
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	api := app.Group("/api/v1")

	// Sprint 12 (ADR 0006): stricter limit on the public, unauthenticated
	// auth endpoints — the most realistic brute-force/credential-stuffing
	// target even at personal scale.
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
	})

	// Sprint 14 sisipan (ADR 0012): credential login endpoints rate-limited
	// tighter than the general auth limiter — brute-force is the most
	// realistic threat against username/password now that it's the
	// default login path.
	credentialLoginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
	})

	authGroup := api.Group("/auth")
	authGroup.Get("/config", authHandler.Config)
	authGroup.Get("/google", authLimiter, authHandler.GoogleLogin)
	authGroup.Get("/google/callback", authLimiter, authHandler.GoogleCallback)
	authGroup.Post("/login", credentialLoginLimiter, authHandler.Login)
	authGroup.Post("/login/admin", credentialLoginLimiter, authHandler.LoginAdmin)
	authGroup.Post("/refresh", authLimiter, authHandler.Refresh)
	authGroup.Post("/logout", requireAuth, authHandler.Logout)
	authGroup.Post("/logout-all", requireAuth, requireOwner, authHandler.LogoutAll)
	authGroup.Get("/me", requireAuth, authHandler.Me)
	authGroup.Put("/me", requireAuth, authHandler.UpdateMe)

	api.Get("/devices", requireAuth, maintenanceGate, deviceHandler.List)
	api.Delete("/devices/:id", requireAuth, maintenanceGate, deviceHandler.Delete)

	ingestGroup := api.Group("/ingest", requireAuth, maintenanceGate)
	ingestGroup.Post("/upload", ingestHandler.Upload)
	ingestGroup.Get("/jobs", ingestHandler.List)
	ingestGroup.Get("/jobs/:id", ingestHandler.Get)
	ingestGroup.Post("/jobs/:id/retry", ingestHandler.Retry)
	ingestGroup.Delete("/jobs/:id", ingestHandler.Delete)

	adminGroup := api.Group("/admin", requireAuth, requireOwner)
	adminGroup.Post("/storage/accounts", storageAccountHandler.Create)
	adminGroup.Get("/storage/accounts", storageAccountHandler.List)
	adminGroup.Delete("/storage/accounts/:id", storageAccountHandler.Delete)
	adminGroup.Post("/storage/accounts/:id/health-check", storageAccountHandler.HealthCheck)
	adminGroup.Post("/ingest-sources/connections", ingestSourceHandler.Connect)
	adminGroup.Get("/ingest-sources/connections", ingestSourceHandler.List)
	adminGroup.Delete("/ingest-sources/connections/:id", ingestSourceHandler.Disconnect)
	adminGroup.Post("/ingest-sources/connections/:id/sync", ingestSourceHandler.Sync)
	adminGroup.Get("/analytics/top-played", analyticsHandler.TopPlayed)
	adminGroup.Get("/analytics/storage-growth", analyticsHandler.StorageGrowth)
	adminGroup.Post("/backup/run", backupHandler.Run)
	adminGroup.Get("/dashboard", dashboardHandler.Get)
	adminGroup.Get("/jobs", ingestHandler.AdminList)
	adminGroup.Post("/jobs/:id/retry", ingestHandler.AdminRetry)
	adminGroup.Get("/lyrics-providers", lyricsProviderHandler.List)
	adminGroup.Patch("/lyrics-providers/:id", lyricsProviderHandler.Update)
	adminGroup.Get("/ingest-sources/:source_type/filters", ingestFilterHandler.List)
	adminGroup.Post("/ingest-sources/:source_type/filters", ingestFilterHandler.Create)
	adminGroup.Delete("/ingest-sources/:source_type/filters/:id", ingestFilterHandler.Delete)
	adminGroup.Get("/users", usersHandler.List)
	adminGroup.Post("/users/invite", usersHandler.Invite)
	adminGroup.Post("/users", usersHandler.Create)
	adminGroup.Delete("/users/:id", usersHandler.Delete)
	adminGroup.Get("/songs", adminSongsHandler.List)
	adminGroup.Patch("/songs/:id", adminSongsHandler.Update)
	adminGroup.Delete("/songs/:id", adminSongsHandler.Delete)
	adminGroup.Get("/settings", adminSettingsHandler.List)
	adminGroup.Patch("/settings", adminSettingsHandler.Update)

	api.Get("/songs/:id", requireAuth, maintenanceGate, catalogHandler.GetSong)
	api.Post("/songs/:id/stream-token", requireAuth, maintenanceGate, catalogHandler.StreamToken)
	// No requireAuth: the browser <audio> element can't send a custom
	// Authorization header, so this route is guarded by the stream token
	// query param instead (ADR 0001).
	api.Get("/songs/:id/stream", catalogHandler.Stream)
	api.Get("/albums/:id", requireAuth, maintenanceGate, catalogHandler.GetAlbum)
	api.Get("/artists/:id", requireAuth, maintenanceGate, catalogHandler.GetArtist)
	api.Get("/artists/:id/albums", requireAuth, maintenanceGate, catalogHandler.ListArtistAlbums)
	api.Get("/artists/:id/songs", requireAuth, maintenanceGate, catalogHandler.ListArtistSongs)
	api.Get("/genres", requireAuth, maintenanceGate, catalogHandler.ListGenres)

	searchGroup := api.Group("/search", requireAuth, maintenanceGate)
	searchGroup.Get("", searchHandler.Search)
	searchGroup.Get("/autocomplete", searchHandler.Autocomplete)
	searchGroup.Get("/trending", searchHandler.Trending)

	playlistGroup := api.Group("/playlists", requireAuth, maintenanceGate)
	playlistGroup.Post("", playlistHandler.Create)
	playlistGroup.Get("", playlistHandler.List)
	playlistGroup.Get("/:id", playlistHandler.Get)
	playlistGroup.Patch("/:id", playlistHandler.Update)
	playlistGroup.Delete("/:id", playlistHandler.Delete)
	playlistGroup.Post("/:id/songs", playlistHandler.AddSong)
	playlistGroup.Patch("/:id/songs/:song_row_id", playlistHandler.UpdateSongPosition)
	playlistGroup.Delete("/:id/songs/:song_row_id", playlistHandler.RemoveSong)

	favoriteGroup := api.Group("/favorites", requireAuth, maintenanceGate)
	favoriteGroup.Get("", favoriteHandler.List)
	favoriteGroup.Post("", favoriteHandler.Create)
	favoriteGroup.Delete("", favoriteHandler.Delete)

	api.Get("/songs/:id/lyrics", requireAuth, maintenanceGate, lyricsHandler.Get)

	historyGroup := api.Group("/history", requireAuth, maintenanceGate)
	historyGroup.Get("", historyHandler.List)
	historyGroup.Post("", historyHandler.Create)
	api.Get("/library/continue-listening", requireAuth, maintenanceGate, historyHandler.ContinueListening)
	api.Get("/library/songs", requireAuth, maintenanceGate, catalogHandler.ListLibrarySongs)
	api.Get("/library/albums", requireAuth, maintenanceGate, catalogHandler.ListLibraryAlbums)
	api.Get("/library/artists", requireAuth, maintenanceGate, catalogHandler.ListLibraryArtists)

	queueGroup := api.Group("/queue", requireAuth, maintenanceGate)
	queueGroup.Get("", queueHandler.List)
	queueGroup.Post("", queueHandler.Add)
	queueGroup.Patch("/:id", queueHandler.UpdatePosition)
	queueGroup.Delete("/:id", queueHandler.Remove)
	queueGroup.Delete("", queueHandler.Clear)

	api.Post("/ws/token", requireAuth, maintenanceGate, wsHandler.IssueToken)
	// No requireAuth: the WS handshake can't send a custom Authorization
	// header, so it's guarded by the single-use ws-token query param
	// (consumed in UpgradeGate) instead — same reasoning as the stream
	// endpoint (ADR 0001).
	app.Get("/ws", wsHandler.UpgradeGate, websocket.New(wsHandler.Handle))

	playerGroup := api.Group("/player", requireAuth, maintenanceGate)
	playerGroup.Get("/state", playerHandler.GetState)
	playerGroup.Post("/state", playerHandler.UpdateState)
	playerGroup.Post("/transfer", playerHandler.Transfer)

	log.Fatal(app.Listen(":8080"))
}
