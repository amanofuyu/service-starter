package redis_test

import (
	"context"
	"testing"
	"time"

	redisstore "service-starter/service/internal/redis"
)

func TestNewClientDoesNotRequireReachableRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client, err := redisstore.NewClient(ctx, "redis://:secret@127.0.0.1:1/0")
	if err != nil {
		t.Fatalf("NewClient returned error for unreachable redis: %v", err)
	}
	defer client.Close()
}
