package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ptop/internal/models"
)

func TestChatCache(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cache := NewChatCache(rdb, 3)

	ctx := context.Background()
	for i := 1; i <= 4; i++ {
		msg := models.OrderMessage{ID: fmt.Sprintf("m%d", i), Content: fmt.Sprintf("%d", i)}
		if err := cache.AddMessage(ctx, "chat1", msg); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}

	history, err := cache.GetHistory(ctx, "chat1")
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(history))
	}
	for idx, want := range []string{"m2", "m3", "m4"} {
		if history[idx].ID != want {
			t.Fatalf("want id %s at %d, got %s", want, idx, history[idx].ID)
		}
	}
}
