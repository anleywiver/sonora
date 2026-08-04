package main

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
)

func main() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
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

	log.Println("sonora-worker starting, connecting to redis:", redisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker failed to start: %v", err)
	}
}
