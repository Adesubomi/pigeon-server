package redis

import (
	"context"

	"github.com/adesubomi/pigeon-server/config"
	redisclient "github.com/redis/go-redis/v9"
)

type Client struct {
	Redis *redisclient.Client
}

func Connect(ctx context.Context, cfg *config.Config) (*Client, error) {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Client{Redis: client}, nil
}

func (c *Client) Close() error {
	if c == nil || c.Redis == nil {
		return nil
	}
	return c.Redis.Close()
}
