package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache provides caching functionality using Redis
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, defaultTTL time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    defaultTTL,
	}
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache with the default TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.ttl)
}

// SetWithTTL stores a value in cache with a custom TTL
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}
	
	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}
	
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) bool {
	result, err := c.client.Exists(ctx, key).Result()
	return err == nil && result > 0
}

// Increment increments a counter in cache
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	
	// Set expiry on first increment
	c.client.Expire(ctx, key, c.ttl)
	
	return val, nil
}

// GetOrSet retrieves from cache or sets if not found
func (c *RedisCache) GetOrSet(ctx context.Context, key string, dest interface{}, 
	fetchFunc func() (interface{}, error)) error {
	
	// Try to get from cache first
	if err := c.Get(ctx, key, dest); err == nil {
		return nil // Found in cache
	}
	
	// Not in cache, fetch fresh data
	data, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}
	
	// Store in cache
	if err := c.Set(ctx, key, data); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache data: %v\n", err)
	}
	
	// Marshal and unmarshal to ensure dest has the right type
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, dest)
	
	return nil
}

// CacheKey generates a cache key with namespace
func CacheKey(namespace string, parts ...string) string {
	key := namespace
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// Cache key namespaces
const (
	ProjectNamespace    = "project"
	UserNamespace       = "user"
	EventsNamespace     = "events"
	AnalyticsNamespace  = "analytics"
	EvaluationNamespace = "evaluation"
	MetricsNamespace    = "metrics"
)

// Cache TTL durations
const (
	ShortTTL  = 1 * time.Minute
	MediumTTL = 5 * time.Minute
	LongTTL   = 15 * time.Minute
	DayTTL    = 24 * time.Hour
)