# Login Unit Test with Redis Mock - Implementation Guide

## Current Status

The Login unit tests have been partially implemented in `service_test.go`, but full Redis mocking is currently **not possible** without refactoring the `AuthService`.

## The Problem

The `AuthService.Login()` method uses `cache.GetRedis()` which returns a global singleton Redis client. This makes it impossible to inject a mock Redis client for unit testing.

```go
// In service.go - Line ~230
redisClient := cache.GetRedis()  // Global singleton - can't be mocked!
```

## Current Test Implementation

The test file includes:
1. ✅ Test for user not found
2. ✅ Test for incorrect password
3. ⚠️ Test for unverified account (skipped - needs Redis mock)
4. ⚠️ Test for successful login (skipped - needs Redis mock)

## Recommended Solution: Dependency Injection

### Step 1: Define Redis Interface

Create a Redis interface in the service package or a shared location:

```go
// internal/domain/authService/service/redis_interface.go
package service

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
}
```

### Step 2: Refactor AuthService Struct

```go
// internal/domain/authService/service/service.go
type AuthService struct {
	userRepo        repository.UserRepository
	userSessionRepo repository.UserSessionRepository
	userProfileRepo repository.UserProfileRepository
	redisClient     RedisClient  // Add this field
}

func NewAuthService(
	userRepo repository.UserRepository,
	userProfileRepo repository.UserProfileRepository,
	userSessionRepo repository.UserSessionRepository,
	redisClient RedisClient,  // Add this parameter
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userSessionRepo: userSessionRepo,
		redisClient:     redisClient,  // Store it
	}
}
```

### Step 3: Update Login Method

Replace all `cache.GetRedis()` calls with `s.redisClient`:

```go
// Before:
redisClient := cache.GetRedis()

// After:
redisClient := s.redisClient
```

### Step 4: Update All Callers

Update all places where `NewAuthService` is called:

```go
// In router setup or handler initialization
authService := service.NewAuthService(
	userRepo,
	userProfileRepo,
	userSessionRepo,
	cache.GetRedis(),  // Pass the real Redis client
)
```

### Step 5: Implement Full Test with Redis Mock

```go
func TestAuthService_Login_Success_WithRedisMock(t *testing.T) {
	db := setupAuthTestDB(t)
	mockUserRepo := new(mocks.MockUserRepository)
	mockProfileRepo := new(mocks.MockUserProfileRepository)
	mockSessionRepo := new(mocks.MockUserSessionRepository)
	mockRedis := new(mocks.MockRedisClient)
	
	// Now we can pass the mock Redis client!
	service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, mockRedis)

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := tokenutils.HashString(password)

	req := dto.LoginRequest{
		Email:     "user@example.com",
		Password:  password,
		IPAddress: "127.0.0.1",
		UserAgent: "Test User Agent",
	}

	existingUser := &model.User{
		ID:         1,
		Email:      req.Email,
		Username:   "testuser",
		FullName:   "Test User",
		Password:   &hashedPassword,
		Provider:   model.ProviderEmail,
		IsVerified: true,
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	createdSession := &model.UserSession{
		ID:             1,
		UserID:         existingUser.ID,
		RefreshTokenID: "test-refresh-token-id",
		IsActive:       true,
	}
	mockSessionRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), 
		mock.AnythingOfType("*model.UserSession")).Return(createdSession, nil)

	// Mock Redis Set for refresh token
	mockSetCmd := redis.NewStatusCmd(context.Background())
	mockSetCmd.SetVal("OK")
	mockRedis.On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "refresh_token:")
	}), mock.Anything, mock.AnythingOfType("time.Duration")).Return(mockSetCmd)

	// Mock Redis SAdd for user session
	mockSAddCmd := redis.NewIntCmd(context.Background())
	mockSAddCmd.SetVal(1)
	mockRedis.On("SAdd", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "user_session:")
	}), mock.Anything).Return(mockSAddCmd)

	// Mock Redis Expire
	mockExpireCmd := redis.NewBoolCmd(context.Background())
	mockExpireCmd.SetVal(true)
	mockRedis.On("Expire", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "user_session:")
	}), mock.AnythingOfType("time.Duration")).Return(mockExpireCmd)

	// Execute Login
	user, accessToken, refreshToken, err := service.Login(ctx, db, req)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, existingUser.Email, user.Email)
	assert.Equal(t, existingUser.Username, user.Username)
	
	// Verify all mocks were called correctly
	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}
```

## Alternative: Integration Test with Redis

If refactoring is not immediately possible, you can write integration tests:

```go
func TestAuthService_Login_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start a real Redis instance using testcontainers or docker
	// Or use miniredis: https://github.com/alicebob/miniredis
	
	// Example with miniredis:
	s := miniredis.RunT(t)
	defer s.Close()
	
	// Configure cache to use test Redis
	// ... test implementation
}
```

## Benefits of Dependency Injection

1. ✅ **Testability**: Easy to mock Redis for unit tests
2. ✅ **Flexibility**: Can swap Redis implementations
3. ✅ **Clarity**: Dependencies are explicit
4. ✅ **Maintainability**: Easier to understand and modify

## Migration Checklist

- [ ] Create RedisClient interface
- [ ] Add redisClient field to AuthService struct
- [ ] Update NewAuthService constructor
- [ ] Replace cache.GetRedis() calls with s.redisClient
- [ ] Update all NewAuthService call sites
- [ ] Ensure MockRedisClient implements RedisClient interface
- [ ] Write comprehensive unit tests with Redis mocks
- [ ] Run all tests to ensure nothing broke

## Related Files

- `internal/domain/authService/service/service.go` - Service implementation
- `internal/domain/authService/service/service_test.go` - Test file
- `internal/domain/mocks/repository_mocks.go` - MockRedisClient
- `internal/infrastructure/cache/redis.go` - Real Redis client

## Notes

The MockRedisClient in `internal/domain/mocks/repository_mocks.go` is already implemented and ready to use once the refactoring is complete.
