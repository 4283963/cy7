package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type clipRecord struct {
	Content        string `json:"c"`
	BurnAfterRead  bool   `json:"b"`
}

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
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func (c *Client) SaveContent(content string, burnAfterRead bool, expiration time.Duration) (string, error) {
	record := clipRecord{
		Content:       content,
		BurnAfterRead: burnAfterRead,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("failed to marshal record: %w", err)
	}

	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		code := generateCode()
		key := fmt.Sprintf("clip:%s", code)

		ok, err := c.rdb.SetNX(ctx, key, payload, expiration).Result()
		if err != nil {
			return "", fmt.Errorf("failed to save content: %w", err)
		}

		if ok {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

var getAndMaybeBurnScript = redis.NewScript(`
local key = KEYS[1]
local raw = redis.call("GET", key)
if raw == false then
	return {err = "code not found or expired"}
end
local ok, record = pcall(cjson.decode, raw)
if not ok then
	redis.call("DEL", key)
	return {err = "corrupted record, removed"}
end
local content = record["c"] or ""
local burn = record["b"] or false
if burn then
	redis.call("DEL", key)
end
return cjson.encode({c = content, b = burn})
`)

type GetResult struct {
	Content       string
	Burned        bool
}

func (c *Client) GetContent(code string) (GetResult, error) {
	key := fmt.Sprintf("clip:%s", code)
	res, err := getAndMaybeBurnScript.Run(ctx, c.rdb, []string{key}).Result()
	if err != nil {
		return GetResult{}, fmt.Errorf("code not found or expired")
	}

	raw, ok := res.(string)
	if !ok {
		return GetResult{}, fmt.Errorf("unexpected script result type")
	}

	var out struct {
		Content string `json:"c"`
		Burned  bool   `json:"b"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return GetResult{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return GetResult{
		Content: out.Content,
		Burned:  out.Burned,
	}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
