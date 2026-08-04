package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/hibiken/asynq"

	appauth "sonora.dev/go-core/application/auth"
	appcatalog "sonora.dev/go-core/application/catalog"
	apphistory "sonora.dev/go-core/application/history"
	appingest "sonora.dev/go-core/application/ingest"
	appingestsource "sonora.dev/go-core/application/ingestsource"
	applibrary "sonora.dev/go-core/application/library"
	applyrics "sonora.dev/go-core/application/lyrics"
	appplayback "sonora.dev/go-core/application/playback"
	appqueue "sonora.dev/go-core/application/queue"
	appsearch "sonora.dev/go-core/application/search"
	appstorage "sonora.dev/go-core/application/storageaccount"
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

	googleClient := oauth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	jwtIssuer := jwt.NewIssuer(cfg.JWTAccessSecret, accessTokenTTL)

	authService := appauth.NewService(userRepo, deviceRepo, refreshTokenRepo, googleClient, jwtIssuer, refreshTokenTTL)
	authHandler := handlers.NewAuthHandler(authService, cfg.FrontendURL, cfg.AdminURL)
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

	wsHub := ws.NewHub()
	wsTokens := wstoken.NewIssuer(redisClient)
	playbackService := appplayback.NewService(queries)
	wsHandler := handlers.NewWSHandler(wsTokens, wsHub, playbackService)
	playerHandler := handlers.NewPlayerHandler(playbackService, authService, wsHub)

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

	ingestGroup := api.Group("/ingest", requireAuth)
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

	api.Get("/songs/:id", requireAuth, catalogHandler.GetSong)
	api.Post("/songs/:id/stream-token", requireAuth, catalogHandler.StreamToken)
	// No requireAuth: the browser <audio> element can't send a custom
	// Authorization header, so this route is guarded by the stream token
	// query param instead (ADR 0001).
	api.Get("/songs/:id/stream", catalogHandler.Stream)
	api.Get("/albums/:id", requireAuth, catalogHandler.GetAlbum)
	api.Get("/artists/:id", requireAuth, catalogHandler.GetArtist)
	api.Get("/artists/:id/albums", requireAuth, catalogHandler.ListArtistAlbums)
	api.Get("/genres", requireAuth, catalogHandler.ListGenres)

	searchGroup := api.Group("/search", requireAuth)
	searchGroup.Get("", searchHandler.Search)
	searchGroup.Get("/autocomplete", searchHandler.Autocomplete)
	searchGroup.Get("/trending", searchHandler.Trending)

	playlistGroup := api.Group("/playlists", requireAuth)
	playlistGroup.Post("", playlistHandler.Create)
	playlistGroup.Get("", playlistHandler.List)
	playlistGroup.Get("/:id", playlistHandler.Get)
	playlistGroup.Patch("/:id", playlistHandler.Update)
	playlistGroup.Delete("/:id", playlistHandler.Delete)
	playlistGroup.Post("/:id/songs", playlistHandler.AddSong)
	playlistGroup.Patch("/:id/songs/:song_row_id", playlistHandler.UpdateSongPosition)
	playlistGroup.Delete("/:id/songs/:song_row_id", playlistHandler.RemoveSong)

	favoriteGroup := api.Group("/favorites", requireAuth)
	favoriteGroup.Get("", favoriteHandler.List)
	favoriteGroup.Post("", favoriteHandler.Create)
	favoriteGroup.Delete("", favoriteHandler.Delete)

	api.Get("/songs/:id/lyrics", requireAuth, lyricsHandler.Get)

	historyGroup := api.Group("/history", requireAuth)
	historyGroup.Get("", historyHandler.List)
	historyGroup.Post("", historyHandler.Create)
	api.Get("/library/continue-listening", requireAuth, historyHandler.ContinueListening)

	queueGroup := api.Group("/queue", requireAuth)
	queueGroup.Get("", queueHandler.List)
	queueGroup.Post("", queueHandler.Add)
	queueGroup.Patch("/:id", queueHandler.UpdatePosition)
	queueGroup.Delete("/:id", queueHandler.Remove)
	queueGroup.Delete("", queueHandler.Clear)

	api.Post("/ws/token", requireAuth, wsHandler.IssueToken)
	// No requireAuth: the WS handshake can't send a custom Authorization
	// header, so it's guarded by the single-use ws-token query param
	// (consumed in UpgradeGate) instead — same reasoning as the stream
	// endpoint (ADR 0001).
	app.Get("/ws", wsHandler.UpgradeGate, websocket.New(wsHandler.Handle))

	playerGroup := api.Group("/player", requireAuth)
	playerGroup.Get("/state", playerHandler.GetState)
	playerGroup.Post("/state", playerHandler.UpdateState)
	playerGroup.Post("/transfer", playerHandler.Transfer)

	log.Fatal(app.Listen(":8080"))
}
