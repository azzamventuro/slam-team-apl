// Package redis provides an optional Redis client. When REDIS_ADDR is unset or
// the server is unreachable, Initialize returns nil and the app runs without it.
package redis

import (
	"context"
	"time"

	"slam-team-api/internal/config"
	"slam-team-api/pkg/logger"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Initialize connects to Redis if configured. Non-fatal: returns nil on any
// failure so the service degrades gracefully.
func Initialize(cfg config.RedisConfig) *goredis.Client {
	if cfg.Addr == "" {
		return nil
	}
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("redis unavailable, continuing without it", zap.Error(err))
		return nil
	}
	logger.Info("redis connected", zap.String("addr", cfg.Addr))
	return client
}
