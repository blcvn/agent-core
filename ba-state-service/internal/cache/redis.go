package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements state management using Redis
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

// SetState saves a key-value pair for a session
func (c *RedisCache) SetState(ctx context.Context, sessionID, key, value string) error {
	redisKey := fmt.Sprintf("state:%s:%s", sessionID, key)
	return c.client.Set(ctx, redisKey, value, c.ttl).Err()
}

// GetState retrieves a value for a session key
func (c *RedisCache) GetState(ctx context.Context, sessionID, key string) (string, bool, error) {
	redisKey := fmt.Sprintf("state:%s:%s", sessionID, key)
	val, err := c.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// ClearState checks if session keys exist and clears them (Pattern matching/Deletion strategy needed)
// For simplicity, this clears a specific session's known keys or uses SCAN which is expensive.
// Better approach: Store keys in a set "session_keys:{sessionID}" to track them for deletion.
func (c *RedisCache) ClearState(ctx context.Context, sessionID string) error {
	// Pattern scan is risky for production but okay for MVP if isolated DB
	// Better: Use a Set to track keys for this session
	// For now, let's just delete the specific keys if we knew them.
	// Given the proto just says "ClearSessionState", maybe we just want to clear ALL keys with prefix?

	iter := c.client.Scan(ctx, 0, fmt.Sprintf("state:%s:*", sessionID), 0).Iterator()
	for iter.Next(ctx) {
		err := c.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
