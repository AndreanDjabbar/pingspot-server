package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"pingspot/internal/config"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var redisInstance redis.UniversalClient

func InitRedis(cfg config.RedisConfig) error {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	var rdb redis.UniversalClient

	if env.RedisHost() != "localhost" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
			TLSConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true,
			},
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   cfg.DB,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis successfully")
	redisInstance = rdb
	return nil
}

func GetRedis() redis.UniversalClient {
	return redisInstance
}
