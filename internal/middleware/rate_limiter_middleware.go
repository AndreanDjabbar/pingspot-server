package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"pingspot/internal/infrastructure/cache"
	"pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	mainutils "pingspot/pkg/utils/mainUtils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimiterConfig struct {
	MaxRequests int
	Window      time.Duration
	KeyPrefix   string
}

type RateLimiter struct {
	redis  redis.UniversalClient
	config RateLimiterConfig
}

func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		redis:  cache.GetRedis(),
		config: config,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, int, error) {
	key := fmt.Sprintf("%s:%s", identifier, rl.config.KeyPrefix,)
	now := time.Now().UnixNano()
	windowStart := now - int64(rl.config.Window.Nanoseconds())

	pipe := rl.redis.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	// remove old entries outside the window

	uniqueUUID := uuid.New().String()

	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d:%s", now, uniqueUUID),
	})
	// add current request timestamp

	countCmd := pipe.ZCard(ctx, key)
	// get current count of requests in the window

	pipe.Expire(ctx, key, rl.config.Window+time.Minute)
	// set expiration for the key

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Rate limiter redis error", zap.Error(err))
		return false, 0, err
	}

	count := int(countCmd.Val())

	if count > rl.config.MaxRequests {
		return false, count, nil
	}

	return true, count, nil
}

func (rl *RateLimiter) GetRemainingRequests(ctx context.Context, identifier string) (int, error) {
	key := fmt.Sprintf("%s:%s", rl.config.KeyPrefix, identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rl.config.Window.Seconds())

	pipe := rl.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	used := int(countCmd.Val())
	remaining := rl.config.MaxRequests - used
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

func GlobalRateLimiterMiddleware() fiber.Handler {
	maxRequest, err := mainutils.StringToInt(env.GlobalRateLimiterMaxRequests())
	if err != nil {
		maxRequest = 500
	}
	windowSeconds, err := mainutils.StringToInt(env.GlobalRateLimiterWindowSeconds())
	if err != nil {
		windowSeconds = 5
	}
	limiter := NewRateLimiter(RateLimiterConfig{
		MaxRequests: maxRequest,
		Window:      time.Duration(windowSeconds) * time.Second,
		KeyPrefix:   "global",
	})

	return func(c *fiber.Ctx) error {
		allowed, count, err := limiter.Allow(c.Context(), "rate_limit:global")
		if err != nil {
			logger.Error("Rate limiter error", zap.Error(err))
			return c.Next()
		}

		c.Set("X-RateLimit-Limit", strconv.Itoa(limiter.config.MaxRequests))
		// Set rate limit headers

		c.Set("X-RateLimit-Remaining", strconv.Itoa(limiter.config.MaxRequests-count))
		// Set remaining requests header

		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.config.Window).Unix(), 10))
		// Set rate limit reset time header

		if !allowed {
			logger.Warn("Rate limit exceeded",
				zap.Int("count", count),
				zap.Int("limit", limiter.config.MaxRequests),
			)
			return apperror.New(
				fiber.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %s allowed.",
					limiter.config.MaxRequests,
					limiter.config.Window),
			)
		}
		return c.Next()
	}
}

func UserRateLimiterMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		if userID == nil {
			userID = mainutils.GetClientIP(c)
		}

		identifier := fmt.Sprintf("rate_limit:user:%v", userID)

		allowed, count, err := limiter.Allow(c.Context(), identifier)
		if err != nil {
			logger.Error("User rate limiter error", zap.Error(err), zap.String("identifier", identifier))
			return c.Next()
		}

		if !allowed {
			logger.Warn("User rate limit exceeded",
				zap.String("identifier", identifier),
				zap.Int("count", count),
				zap.Int("limit", limiter.config.MaxRequests),
			)
			return apperror.New(
				fiber.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %s allowed.",
					limiter.config.MaxRequests,
					limiter.config.Window),
			)
		}

		c.Set("X-RateLimit-Limit", strconv.Itoa(limiter.config.MaxRequests))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(limiter.config.MaxRequests-count))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.config.Window).Unix(), 10))

		return c.Next()
	}
}