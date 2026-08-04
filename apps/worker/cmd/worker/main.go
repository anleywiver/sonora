package main

import (
	"context"
	"log"

	"github.com/hibiken/asynq"

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
	// Task handler akan didaftarkan di sini mulai Sprint 3 (ingest pipeline)
	// mux.HandleFunc("ingest:process", tasks.HandleIngestTask)

	log.Println("sonora-worker starting, connecting to redis:", cfg.RedisURL)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker failed to start: %v", err)
	}
}
