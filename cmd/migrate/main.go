package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"SmartFeed/internal/config"
	"SmartFeed/migrations"
	"SmartFeed/pkg/postgres"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: go run ./cmd/migrate up|down")
	}
	command := os.Args[1]

	cfg := config.Load()
	pool, err := postgres.New(context.Background(), cfg.PGDSN)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect error: %v", err)
	}

	switch command {
	case "up":
		if err := goose.Up(db, "."); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
	case "down":
		if err := goose.Down(db, "."); err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
	default:
		log.Fatalf("unknown command %q, expected up|down", command)
	}

	log.Printf("migration command %q done", command)
}
