package database_test

import (
	"context"
	"testing"
	"time"

	"service-starter/service/internal/database"
)

func TestNewPoolDoesNotRequireReachableDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	pool, err := database.NewPool(ctx, "postgres://service:secret@127.0.0.1:1/service?sslmode=disable")
	if err != nil {
		t.Fatalf("NewPool returned error for unreachable database: %v", err)
	}
	defer pool.Close()
}
