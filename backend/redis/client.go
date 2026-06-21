package redis

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr, password string, db int) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func generateCode() string {
	return fmt.Sprintf("%04d", rng.Intn(10000))
}

func (c *Client) SaveContent(content string, expiration time.Duration) (string, error) {
	var code string
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		code = generateCode()
		key := fmt.Sprintf("clip:%s", code)

		exists, err := c.rdb.Exists(ctx, key).Result()
		if err != nil {
			return "", fmt.Errorf("failed to check key existence: %w", err)
		}

		if exists == 0 {
			err = c.rdb.Set(ctx, key, content, expiration).Err()
			if err != nil {
				return "", fmt.Errorf("failed to save content: %w", err)
			}
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

func (c *Client) GetContent(code string) (string, error) {
	key := fmt.Sprintf("clip:%s", code)
	content, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("code not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get content: %w", err)
	}
	return content, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
