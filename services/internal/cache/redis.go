package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"articleswap-scalable/services/internal/config"
)

// RedisCache wraps the Redis client with helper methods for caching.
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache creates a new Redis client and verifies connectivity.
func NewRedisCache(ctx context.Context, cfg *config.Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("[redis] connected — addr=%s db=%d ttl=%s",
		cfg.RedisAddr, cfg.RedisDB, cfg.CacheTTL)

	return &RedisCache{
		client: client,
		ttl:    cfg.CacheTTL,
	}, nil
}

// Get retrieves a value from cache. Returns empty string and nil error if key
// does not exist (cache miss).
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a value in cache with the default TTL.
func (r *RedisCache) Set(ctx context.Context, key, value string) error {
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

// SetWithTTL stores a value in cache with a custom TTL.
func (r *RedisCache) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Incr increments a counter and returns the new value.
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire sets a TTL on an existing key.
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// Client exposes the underlying Redis client for advanced operations.
func (r *RedisCache) Client() *redis.Client {
	return r.client
}

// HealthCheck pings Redis to verify connectivity.
func (r *RedisCache) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close shuts down the Redis client.
func (r *RedisCache) Close() error {
	return r.client.Close()
}
