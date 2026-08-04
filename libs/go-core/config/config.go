// Package config loads process configuration from environment variables.
// Shared by apps/backend and apps/worker so both read the same env contract.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string

	MeilisearchURL    string
	MeilisearchAPIKey string

	JWTAccessSecret  string
	JWTRefreshSecret string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	StorageCredentialsEncryptionKey string

	// FrontendURL is where the browser gets redirected after the OAuth
	// callback finishes (access token delivered as a URL fragment).
	FrontendURL string
}

// Load reads .env (if present — Docker Compose injects env vars directly and
// won't have the file, so a missing .env is not an error) then env vars.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:                     os.Getenv("DATABASE_URL"),
		RedisURL:                        os.Getenv("REDIS_URL"),
		MeilisearchURL:                  os.Getenv("MEILISEARCH_URL"),
		MeilisearchAPIKey:               os.Getenv("MEILISEARCH_API_KEY"),
		JWTAccessSecret:                 os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:                os.Getenv("JWT_REFRESH_SECRET"),
		GoogleClientID:                  os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:              os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:               os.Getenv("GOOGLE_REDIRECT_URL"),
		StorageCredentialsEncryptionKey: os.Getenv("STORAGE_CREDENTIALS_ENCRYPTION_KEY"),
		FrontendURL:                     envOrDefault("FRONTEND_URL", "http://localhost:3000"),
	}

	required := map[string]string{
		"DATABASE_URL":       cfg.DatabaseURL,
		"REDIS_URL":          cfg.RedisURL,
		"JWT_ACCESS_SECRET":  cfg.JWTAccessSecret,
		"JWT_REFRESH_SECRET": cfg.JWTRefreshSecret,
	}
	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("config: required env var %s is not set", name)
		}
	}

	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return fallback
}
