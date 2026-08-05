// Command seed-owner bootstraps the very first Owner account for
// credential-based login (Sprint 14 sisipan, ADR 0012). Deliberately a
// manual CLI, not an HTTP endpoint — an endpoint that can create an
// Owner would let anyone who reaches the API before bootstrap make
// themselves Owner.
//
// Usage:
//
//	go run ./cmd/seed-owner --email="owner@example.com" --password="a strong password"
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"

	"sonora.dev/go-core/config"
	"sonora.dev/go-core/domain/identity"
	"sonora.dev/go-core/infrastructure/passwordhash"
	"sonora.dev/go-core/infrastructure/postgres"
	"sonora.dev/go-core/infrastructure/postgres/repository"
)

func main() {
	email := flag.String("email", "", "Owner's login email (required)")
	password := flag.String("password", "", "Owner's login password (required, min 8 chars)")
	name := flag.String("name", "Owner", "display name")
	flag.Parse()

	if *email == "" || *password == "" {
		log.Fatal("seed-owner: --email and --password are required")
	}
	if len(*password) < 8 {
		log.Fatal("seed-owner: password must be at least 8 characters")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("seed-owner: config: %v", err)
	}

	ctx := context.Background()
	gormDB, err := postgres.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("seed-owner: gorm: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("seed-owner: gorm underlying db: %v", err)
	}
	defer sqlDB.Close()

	userRepo := repository.NewUserRepository(gormDB)

	if _, err := userRepo.FindByEmail(ctx, *email); err == nil {
		log.Fatalf("seed-owner: a user with email %q already exists", *email)
	} else if !errors.Is(err, identity.ErrNotFound) {
		log.Fatalf("seed-owner: check existing email: %v", err)
	}

	hash, err := passwordhash.Hash(*password)
	if err != nil {
		log.Fatalf("seed-owner: hash password: %v", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("seed-owner: generate id: %v", err)
	}

	user := &identity.User{
		ID:           id,
		Email:        *email,
		Name:         *name,
		Role:         identity.RoleOwner,
		PasswordHash: hash,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		log.Fatalf("seed-owner: create: %v", err)
	}

	fmt.Printf("Owner created: id=%s email=%s\n", user.ID, user.Email)
}
