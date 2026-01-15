package middleware

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRedisClient creates a mock Redis client for testing
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use test database
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ping to check connection
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	// Clear test database
	client.FlushDB(ctx)

	return client
}

func TestRateLimiter_Allow(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	tests := []struct {
		name          string
		config        RateLimiterConfig
		requests      int
		identifier    string
		expectAllowed []bool
	}{
		{
			name: "allows requests within limit",
			config: RateLimiterConfig{
				MaxRequests: 5,
				Window:      1 * time.Minute,
				KeyPrefix:   "test:limit1",
			},
			requests:      5,
			identifier:    "user123",
			expectAllowed: []bool{true, true, true, true, true},
		},
		{
			name: "blocks requests exceeding limit",
			config: RateLimiterConfig{
				MaxRequests: 3,
				Window:      1 * time.Minute,
				KeyPrefix:   "test:limit2",
			},
			requests:      5,
			identifier:    "user456",
			expectAllowed: []bool{true, true, true, false, false},
		},
		{
			name: "different identifiers have separate limits",
			config: RateLimiterConfig{
				MaxRequests: 2,
				Window:      1 * time.Minute,
				KeyPrefix:   "test:limit3",
			},
			requests:      2,
			identifier:    "user789",
			expectAllowed: []bool{true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear Redis before each test
			ctx := context.Background()
			key := fmt.Sprintf("%s:%s", tt.config.KeyPrefix, tt.identifier)
			client.Del(ctx, key)

			limiter := &RateLimiter{
				redis:  client,
				config: tt.config,
			}

			for i := 0; i < tt.requests; i++ {
				allowed, count, err := limiter.Allow(ctx, tt.identifier)
				require.NoError(t, err)

				if i < len(tt.expectAllowed) {
					assert.Equal(t, tt.expectAllowed[i], allowed,
						"Request %d: expected allowed=%v, got %v", i+1, tt.expectAllowed[i], allowed)
				}

				if allowed {
					assert.Equal(t, i+1, count, "Count mismatch at request %d", i+1)
				}
			}

			// Clean up
			client.Del(ctx, key)
		})
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	config := RateLimiterConfig{
		MaxRequests: 2,
		Window:      2 * time.Second,
		KeyPrefix:   "test:expiry",
	}

	limiter := &RateLimiter{
		redis:  client,
		config: config,
	}

	ctx := context.Background()
	identifier := "user_expiry"

	// Make 2 requests (should succeed)
	allowed1, _, err := limiter.Allow(ctx, identifier)
	require.NoError(t, err)
	assert.True(t, allowed1)

	allowed2, _, err := limiter.Allow(ctx, identifier)
	require.NoError(t, err)
	assert.True(t, allowed2)

	// Third request should be blocked
	allowed3, _, err := limiter.Allow(ctx, identifier)
	require.NoError(t, err)
	assert.False(t, allowed3)

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// Should be allowed again after window expires
	allowed4, _, err := limiter.Allow(ctx, identifier)
	require.NoError(t, err)
	assert.True(t, allowed4)

	// Clean up
	key := fmt.Sprintf("%s:%s", config.KeyPrefix, identifier)
	client.Del(ctx, key)
}

func TestRateLimiter_GetRemainingRequests(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	config := RateLimiterConfig{
		MaxRequests: 5,
		Window:      1 * time.Minute,
		KeyPrefix:   "test:remaining",
	}

	limiter := &RateLimiter{
		redis:  client,
		config: config,
	}

	ctx := context.Background()
	identifier := "user_remaining"

	// Initially should have full limit available
	remaining, err := limiter.GetRemainingRequests(ctx, identifier)
	require.NoError(t, err)
	assert.Equal(t, 5, remaining)

	// Make 2 requests
	limiter.Allow(ctx, identifier)
	limiter.Allow(ctx, identifier)

	// Should have 3 remaining
	remaining, err = limiter.GetRemainingRequests(ctx, identifier)
	require.NoError(t, err)
	assert.Equal(t, 3, remaining)

	// Make 3 more requests
	limiter.Allow(ctx, identifier)
	limiter.Allow(ctx, identifier)
	limiter.Allow(ctx, identifier)

	// Should have 0 remaining
	remaining, err = limiter.GetRemainingRequests(ctx, identifier)
	require.NoError(t, err)
	assert.Equal(t, 0, remaining)

	// Clean up
	key := fmt.Sprintf("%s:%s", config.KeyPrefix, identifier)
	client.Del(ctx, key)
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skip("Redis not available for benchmarking")
	}
	defer client.Close()

	config := RateLimiterConfig{
		MaxRequests: 1000,
		Window:      1 * time.Minute,
		KeyPrefix:   "benchmark:limit",
	}

	limiter := &RateLimiter{
		redis:  client,
		config: config,
	}

	identifier := "benchmark_user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, identifier)
	}

	// Clean up
	key := fmt.Sprintf("%s:%s", config.KeyPrefix, identifier)
	client.Del(ctx, key)
}
