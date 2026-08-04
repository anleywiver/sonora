package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	appingest "sonora.dev/go-core/application/ingest"
	appingestsource "sonora.dev/go-core/application/ingestsource"
	appmaintenance "sonora.dev/go-core/application/maintenance"
	appstorage "sonora.dev/go-core/application/storageaccount"
	"sonora.dev/go-core/config"
	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/meilisearch"
	"sonora.dev/go-core/infrastructure/postgres"
	"sonora.dev/go-core/infrastructure/postgres/repository"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
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

	credentialsBox, err := crypto.NewBox(cfg.StorageCredentialsEncryptionKey)
	if err != nil {
		log.Fatalf("storage credentials box: %v", err)
	}
	queries := sqlc.New(pool)
	meiliClient := meilisearch.NewClient(cfg.MeilisearchURL, cfg.MeilisearchAPIKey)
	ingestService := appingest.NewService(queries, credentialsBox, meiliClient, cfg.GoogleClientID, cfg.GoogleClientSecret)

	userRepo := repository.NewUserRepository(gormDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(gormDB)
	storageAccountService := appstorage.NewService(queries, credentialsBox, cfg.GoogleClientID, cfg.GoogleClientSecret)
	ingestSourceService := appingestsource.NewService(queries, credentialsBox, userRepo, ingestService, cfg.IngestTmpDir)
	maintenanceService := appmaintenance.NewService(queries, refreshTokenRepo, storageAccountService)

	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisURL}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Concurrency dibatasi supaya tidak overload storage provider
			// (lihat concern rate-limit Google Drive di STEP 3 System Flow)
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(appingest.TaskTypeProcess, func(ctx context.Context, t *asynq.Task) error {
		var payload appingest.ProcessPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("ingest:process: unmarshal payload: %w", err)
		}
		return ingestService.Process(ctx, payload.JobID)
	})
	// Sprint 10: these three carry no payload — the scheduler enqueues them
	// empty, and the handlers loop internally best-effort (see
	// application/maintenance and application/ingestsource).
	mux.HandleFunc(appmaintenance.TaskTypeGarbageCollect, func(ctx context.Context, t *asynq.Task) error {
		maintenanceService.GarbageCollect(ctx)
		return nil
	})
	mux.HandleFunc(appmaintenance.TaskTypeStorageOptimize, func(ctx context.Context, t *asynq.Task) error {
		maintenanceService.StorageOptimize(ctx)
		return nil
	})
	mux.HandleFunc(appingestsource.TaskTypeSyncAll, func(ctx context.Context, t *asynq.Task) error {
		ingestSourceService.SyncAll(ctx)
		return nil
	})

	// Sprint 10: Asynq Scheduler runs as a goroutine in this same process
	// rather than a separate deployable — no reason for a third process on
	// a 1-VPS deployment just for cron (ADR 0004).
	scheduler := asynq.NewScheduler(redisOpt, nil)
	if _, err := scheduler.Register("0 3 * * *", asynq.NewTask(appmaintenance.TaskTypeGarbageCollect, nil)); err != nil {
		log.Fatalf("schedule garbage collector: %v", err)
	}
	if _, err := scheduler.Register("0 4 * * 0", asynq.NewTask(appmaintenance.TaskTypeStorageOptimize, nil)); err != nil {
		log.Fatalf("schedule storage optimizer: %v", err)
	}
	if _, err := scheduler.Register("0 */6 * * *", asynq.NewTask(appingestsource.TaskTypeSyncAll, nil)); err != nil {
		log.Fatalf("schedule ingest source sync: %v", err)
	}
	go func() {
		if err := scheduler.Run(); err != nil {
			log.Fatalf("scheduler failed to start: %v", err)
		}
	}()

	log.Println("sonora-worker starting, connecting to redis:", cfg.RedisURL)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker failed to start: %v", err)
	}
}
