# Rate Limiter Usage Guide

## Overview

The rate limiter provides three middleware options for protecting your API endpoints from excessive requests:

1. **Global Rate Limiter** - Limits requests by IP address globally
2. **User Rate Limiter** - Limits requests per authenticated user
3. **Endpoint Rate Limiter** - Limits requests per specific endpoint

## Features

- ✅ Redis-backed for distributed rate limiting
- ✅ Sliding window algorithm for accurate rate limiting
- ✅ Automatic cleanup of expired entries
- ✅ Rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`)
- ✅ Graceful error handling (allows requests on Redis errors)
- ✅ Comprehensive logging

## Usage Examples

### 1. Global Rate Limiter

Apply to all routes to limit requests by IP address:

```go
// In server.go
func (s *FiberServer) RegisterFiberRoutes() {
    // Apply global rate limiter to all routes
    s.App.Use(middleware.GlobalRateLimiterMiddleware())
    
    // Your other routes...
    router.RegisterRoutes(s.App)
}
```

**Default Configuration:**
- 100 requests per minute per IP address
- Returns 429 status code when exceeded

### 2. User Rate Limiter

Apply to authenticated routes to limit per user:

```go
// Create a custom rate limiter for authenticated users
userLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
    MaxRequests: 50,              // 50 requests
    Window:      1 * time.Minute, // per minute
    KeyPrefix:   "rate_limit:user",
})

// Apply to authenticated routes
authRoutes := app.Group("/api/auth")
authRoutes.Use(middleware.UserRateLimiterMiddleware(userLimiter))
```

**Note:** Requires `userID` to be set in `c.Locals("userID")` by your auth middleware.

### 3. Endpoint-Specific Rate Limiter

Apply to specific endpoints that need stricter limits:

```go
// In your router file (e.g., authRouter)
func RegisterAuthRoutes(app *fiber.App) {
    auth := app.Group("/api/auth")
    
    // Strict limit on login endpoint (prevent brute force)
    auth.Post("/login", 
        middleware.EndpointRateLimiterMiddleware(5, 15*time.Minute, "login"),
        authController.Login,
    )
    
    // Strict limit on registration (prevent spam)
    auth.Post("/register",
        middleware.EndpointRateLimiterMiddleware(3, 1*time.Hour, "register"),
        authController.Register,
    )
    
    // Limit password reset requests
    auth.Post("/forgot-password",
        middleware.EndpointRateLimiterMiddleware(3, 30*time.Minute, "forgot-password"),
        authController.ForgotPassword,
    )
}
```

### 4. Combined Rate Limiters

You can combine multiple rate limiters for layered protection:

```go
// Global rate limiter for all routes
s.App.Use(middleware.GlobalRateLimiterMiddleware())

// User-specific limiter for authenticated routes
userLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
    MaxRequests: 200,
    Window:      1 * time.Minute,
    KeyPrefix:   "rate_limit:user",
})

authRoutes := app.Group("/api/auth")
authRoutes.Use(middleware.UserRateLimiterMiddleware(userLimiter))

// Endpoint-specific for sensitive operations
authRoutes.Post("/login",
    middleware.EndpointRateLimiterMiddleware(5, 15*time.Minute, "login"),
    authController.Login,
)
```

## Configuration Options

### RateLimiterConfig

```go
type RateLimiterConfig struct {
    MaxRequests int           // Maximum number of requests
    Window      time.Duration // Time window for rate limiting
    KeyPrefix   string        // Redis key prefix
}
```

### Common Configurations

```go
// Very strict (for sensitive endpoints)
middleware.RateLimiterConfig{
    MaxRequests: 5,
    Window:      15 * time.Minute,
    KeyPrefix:   "rate_limit:strict",
}

// Standard (for most endpoints)
middleware.RateLimiterConfig{
    MaxRequests: 100,
    Window:      1 * time.Minute,
    KeyPrefix:   "rate_limit:standard",
}

// Relaxed (for public endpoints)
middleware.RateLimiterConfig{
    MaxRequests: 500,
    Window:      1 * time.Minute,
    KeyPrefix:   "rate_limit:relaxed",
}
```

## Response Headers

The middleware automatically adds these headers to responses:

- `X-RateLimit-Limit`: Maximum number of requests allowed
- `X-RateLimit-Remaining`: Number of requests remaining in the current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets

Example response headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1705075200
```

## Error Response

When rate limit is exceeded, the API returns:

```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests. Please try again later.",
  "details": "Rate limit exceeded. Maximum 100 requests per 1m0s allowed.",
  "statusCode": 429
}
```

## Best Practices

1. **Layer Your Rate Limiters**: Use global + user + endpoint limiters for comprehensive protection
2. **Adjust Based on Endpoint Sensitivity**:
   - Login/Register: 5 requests / 15 minutes
   - Password Reset: 3 requests / 30 minutes
   - File Upload: 10 requests / 5 minutes
   - Search/Read: 100 requests / 1 minute
   - Write Operations: 50 requests / 1 minute

3. **Monitor Logs**: The rate limiter logs warnings when limits are exceeded
4. **Handle Client-Side**: Check `X-RateLimit-*` headers and implement backoff strategy
5. **Test in Development**: Use shorter windows during testing

## Testing

Run the tests:

```bash
# Make sure Redis is running
make test-rate-limiter

# Or run directly
go test ./internal/middleware -v -run TestRateLimiter

# Benchmark
go test ./internal/middleware -bench=BenchmarkRateLimiter -benchmem
```

## Example Implementation in Your Project

Here's a complete example for your server setup:

```go
// In internal/server/server.go
func (s *FiberServer) RegisterFiberRoutes() {
    s.App.Use(cors.New(cors.Config{
        AllowOrigins:     "http://localhost:3000",
        AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
        AllowHeaders:     "Accept,Authorization,Content-Type",
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Global rate limiter (100 req/min per IP)
    s.App.Use(middleware.GlobalRateLimiterMiddleware())

    defaultRoute := s.App.Group("/pingspot/api")
    defaultRoute.Get("/", DefaultHandler)

    router.RegisterRoutes(s.App)
}
```

```go
// In internal/domain/authService/router/auth_router.go
func RegisterAuthRoutes(app *fiber.App) {
    auth := app.Group("/api/auth")
    
    // Endpoint-specific rate limiters
    auth.Post("/login",
        middleware.EndpointRateLimiterMiddleware(5, 15*time.Minute, "login"),
        authController.Login,
    )
    
    auth.Post("/register",
        middleware.EndpointRateLimiterMiddleware(3, 1*time.Hour, "register"),
        authController.Register,
    )
    
    auth.Post("/forgot-password",
        middleware.EndpointRateLimiterMiddleware(3, 30*time.Minute, "forgot-password"),
        authController.ForgotPassword,
    )
}
```

## Troubleshooting

### Rate Limiter Not Working

1. Check Redis connection:
   ```bash
   redis-cli ping
   ```

2. Verify Redis is initialized in your app:
   ```go
   // In cmd/main.go, ensure Redis is initialized
   cache.InitRedis(config.LoadRedisConfig())
   ```

3. Check logs for errors:
   ```bash
   grep "rate limiter" logs/app.log
   ```

### Testing Without Redis

The middleware gracefully handles Redis errors and allows requests to proceed, ensuring your API remains available even if Redis is down.

## Environment Variables

Make sure these are set in your `.env` file:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
```




func EndpointRateLimiterMiddleware(maxRequests int, window time.Duration, endpoint string) fiber.Handler {
	limiter := NewRateLimiter(RateLimiterConfig{
		MaxRequests: maxRequests,
		Window:      window,
		KeyPrefix:   fmt.Sprintf("rate_limit:endpoint:%s", endpoint),
	})

	return func(c *fiber.Ctx) error {
		identifier := mainutils.GetClientIP(c)
		if userID := c.Locals("userID"); userID != nil {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		allowed, count, err := limiter.Allow(c.Context(), identifier)
		if err != nil {
			logger.Error("Endpoint rate limiter error",
				zap.Error(err),
				zap.String("endpoint", endpoint),
				zap.String("identifier", identifier),
			)
			return c.Next()
		}

		c.Set("X-RateLimit-Limit", strconv.Itoa(limiter.config.MaxRequests))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(limiter.config.MaxRequests-count))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.config.Window).Unix(), 10))

		if !allowed {
			logger.Warn("Endpoint rate limit exceeded",
				zap.String("endpoint", endpoint),
				zap.String("identifier", identifier),
				zap.Int("count", count),
				zap.Int("limit", limiter.config.MaxRequests),
			)
			return apperror.New(
				fiber.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				fmt.Sprintf("Rate limit exceeded for this endpoint. Maximum %d requests per %s allowed.",
					maxRequests,
					window),
			)
		}
		return c.Next()
	}
}
