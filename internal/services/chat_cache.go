package services

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"ptop/internal/models"
)

type ChatCache struct {
	client *redis.Client
	limit  int64
}

func NewChatCache(client *redis.Client, limit int64) *ChatCache {
	return &ChatCache{client: client, limit: limit}
}

func (c *ChatCache) AddMessage(ctx context.Context, chatID string, msg models.OrderMessage) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	key := "chat:" + chatID + ":messages"
	pipe := c.client.TxPipeline()
	pipe.LPush(ctx, key, b)
	pipe.LTrim(ctx, key, 0, c.limit-1)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *ChatCache) GetHistory(ctx context.Context, chatID string) ([]models.OrderMessage, error) {
	key := "chat:" + chatID + ":messages"
	vals, err := c.client.LRange(ctx, key, 0, c.limit-1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	res := make([]models.OrderMessage, 0, len(vals))
	for i := len(vals) - 1; i >= 0; i-- {
		var m models.OrderMessage
		if e := json.Unmarshal([]byte(vals[i]), &m); e == nil {
			res = append(res, m)
		}
	}
	return res, nil
}
