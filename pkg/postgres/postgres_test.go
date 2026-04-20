package postgres

import (
	"context"
	"testing"
)

func TestNewInvalidDSN(t *testing.T) {
	if _, err := New(context.Background(), "::bad-dsn::"); err == nil {
		t.Fatal("expected parse dsn error")
	}
}
