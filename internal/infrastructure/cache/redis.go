package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"pingspot/internal/config"
	"pingspot/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var redisInstance redis.UniversalClient

func InitRedis(cfg config.RedisConfig) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	options := &redis.Options{
		Addr:     addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.UseTLS {
		options.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
	}

	rdb := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully", zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("useTLS", fmt.Sprintf("%v", cfg.UseTLS)))
	redisInstance = rdb
	return nil
}

func GetRedis() redis.UniversalClient {
	return redisInstance
}
