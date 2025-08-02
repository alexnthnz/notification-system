package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/alexnthnz/notification-system/internal/config"
)

// RedisClient wraps redis.Client for caching operations
type RedisClient struct {
	*redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}

// CacheUserPreferences caches user notification preferences
func (r *RedisClient) CacheUserPreferences(ctx context.Context, userID string, preferences interface{}) error {
	key := fmt.Sprintf("user_preferences:%s", userID)
	return r.Set(ctx, key, preferences, time.Hour).Err()
}

// GetUserPreferences retrieves cached user notification preferences
func (r *RedisClient) GetUserPreferences(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user_preferences:%s", userID)
	return r.Get(ctx, key).Result()
}

// CacheNotificationTemplate caches notification templates
func (r *RedisClient) CacheNotificationTemplate(ctx context.Context, templateName string, template interface{}) error {
	key := fmt.Sprintf("template:%s", templateName)
	return r.Set(ctx, key, template, 24*time.Hour).Err()
}

// GetNotificationTemplate retrieves cached notification template
func (r *RedisClient) GetNotificationTemplate(ctx context.Context, templateName string) (string, error) {
	key := fmt.Sprintf("template:%s", templateName)
	return r.Get(ctx, key).Result()
}

// IncrementRateLimit increments rate limit counter for a user
func (r *RedisClient) IncrementRateLimit(ctx context.Context, userID string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("rate_limit:%s", userID)
	pipe := r.Pipeline()
	
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incr.Val(), nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.Client.Close()
}