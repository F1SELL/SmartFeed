package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PG_DSN", "postgres://smartfeed:secret@localhost:55432/smartfeed_db?sslmode=disable")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("JWT_SECRET", "secret")

	cfg := Load()
	if cfg == nil {
		t.Fatal("config is nil")
	}
	if cfg.PGDSN == "" {
		t.Fatal("PG_DSN should not be empty")
	}
	if cfg.JWTSecret == "" {
		t.Fatal("JWT_SECRET should not be empty")
	}
}
