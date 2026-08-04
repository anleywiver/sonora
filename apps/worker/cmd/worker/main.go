package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	appingest "sonora.dev/go-core/application/ingest"
	"sonora.dev/go-core/config"
	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/meilisearch"
	"sonora.dev/go-core/infrastructure/postgres"
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

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisURL},
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

	log.Println("sonora-worker starting, connecting to redis:", cfg.RedisURL)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker failed to start: %v", err)
	}
}
