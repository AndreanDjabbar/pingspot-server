# Unit Testing Guide

This document provides comprehensive information about the unit testing implementation for the PingSpot server.

## Overview

The server now has comprehensive unit test coverage including:
- **Utility Functions**: Token utilities, main utilities
- **Services**: User service, Auth service
- **Error Handling**: AppError package
- **Mock Repositories**: For testing services in isolation

## Test Structure

```
server/
├── pkg/
│   ├── appError/
│   │   ├── error.go
│   │   └── error_test.go
│   └── utils/
│       ├── tokenutils/
│       │   ├── token_util.go
│       │   └── token_util_test.go
│       └── mainUtils/
│           ├── main_util.go
│           └── main_util_test.go
├── internal/
│   └── domain/
│       ├── mocks/
│       │   └── repository_mocks.go
│       ├── authService/
│       │   └── service/
│       │       ├── service.go
│       │       └── service_test.go
│       └── userService/
│           └── service/
│               ├── service.go
│               └── service_test.go
```

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage report
make test-coverage

# Run short tests (skip integration tests)
make test-short

# Run benchmark tests
make test-benchmark

# Run tests with verbose output
make test-verbose

# Watch mode (auto-rerun on changes)
make test-watch
```

### Manual Commands

```bash
# Run all tests
go test ./... -v

# Run tests in a specific package
go test ./pkg/utils/tokenutils -v

# Run specific test
go test ./pkg/utils/tokenutils -run TestHashString -v

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test ./... -bench=. -benchmem
```

## Test Coverage

### Current Coverage

1. **Token Utilities** (`pkg/utils/tokenutils`)
   - ✅ Password hashing and verification (bcrypt)
   - ✅ SHA256 hashing
   - ✅ Random code generation
   - ✅ JWT generation and parsing
   - ✅ Benchmark tests for performance

2. **Main Utilities** (`pkg/utils/mainUtils`)
   - ✅ Client IP extraction from various headers
   - ✅ User agent parsing
   - ✅ Device info formatting
   - ✅ Key path generation
   - ✅ Email template rendering

3. **User Service** (`internal/domain/userService/service`)
   - ✅ Get user profile by ID
   - ✅ Get user profile by username
   - ✅ Update password security
   - ✅ Error handling for not found users
   - ✅ Database error handling

4. **Auth Service** (`internal/domain/authService/service`)
   - ✅ User registration
   - ✅ Email uniqueness validation
   - ✅ Password hashing during registration
   - ✅ Profile creation alongside user
   - ✅ Transaction rollback on errors

5. **AppError Package** (`pkg/appError`)
   - ✅ Error creation with all fields
   - ✅ Error message retrieval
   - ✅ Common error patterns
   - ✅ Special character handling

## Writing Tests

### Test File Naming

Test files should be named with `_test.go` suffix:
- `service.go` → `service_test.go`
- `token_util.go` → `token_util_test.go`

### Test Function Naming

```go
func TestFunctionName(t *testing.T) {
    t.Run("should do something successfully", func(t *testing.T) {
        // Test implementation
    })
}
```

### Example Test Structure

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserService_GetProfile(t *testing.T) {
    t.Run("should get user profile successfully", func(t *testing.T) {
        // Arrange
        mockRepo := new(mocks.MockUserRepository)
        service := NewUserService(mockRepo, nil)
        
        expectedUser := &model.User{
            ID: 1,
            Email: "test@example.com",
        }
        mockRepo.On("GetByID", mock.Anything, uint(1)).Return(expectedUser, nil)
        
        // Act
        result, err := service.GetProfile(context.Background(), 1)
        
        // Assert
        require.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, uint(1), result.UserID)
        mockRepo.AssertExpectations(t)
    })
}
```

## Mock Repositories

Mock repositories are available in `internal/domain/mocks/`:

```go
// Using mock repository
mockUserRepo := new(mocks.MockUserRepository)
mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)
```

### Available Mocks

- `MockUserRepository`
- `MockUserProfileRepository`
- `MockUserSessionRepository`

## Best Practices

### 1. Use Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := Function(tc.input)
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### 2. Test Both Success and Failure Cases

Always test:
- ✅ Happy path (success scenarios)
- ✅ Error cases (validation errors, not found, etc.)
- ✅ Edge cases (empty strings, nil values, etc.)

### 3. Use Descriptive Test Names

```go
// Good
t.Run("should return error when user not found", func(t *testing.T) { ... })

// Bad
t.Run("test1", func(t *testing.T) { ... })
```

### 4. Isolate Tests

Each test should be independent and not rely on other tests.

### 5. Clean Up Resources

```go
func TestFunction(t *testing.T) {
    db := setupTestDB(t)
    sqlDB, _ := db.DB()
    defer sqlDB.Close() // Always close resources
    
    // Test implementation
}
```

## Integration Tests

Integration tests can be skipped in short mode:

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Integration test implementation
}
```

Run only unit tests:
```bash
make test-short
```

## Benchmark Tests

Write benchmarks to measure performance:

```go
func BenchmarkHashString(b *testing.B) {
    password := "mySecretPassword123"
    for i := 0; i < b.N; i++ {
        _, _ = HashString(password)
    }
}
```

Run benchmarks:
```bash
make test-benchmark
```

## Code Coverage Goals

Target coverage by package:
- **Utilities**: 80%+ coverage
- **Services**: 70%+ coverage
- **Handlers**: 60%+ coverage

View current coverage:
```bash
make test-coverage
# Open coverage.html in browser
```

## Continuous Integration

Tests should be run automatically on:
- Every commit
- Pull requests
- Before deployment

## Common Testing Patterns

### Testing with Context

```go
ctx := context.Background()
result, err := service.GetProfile(ctx, userID)
```

### Testing with Mock Database

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    err = db.AutoMigrate(&model.User{})
    require.NoError(t, err)
    
    return db
}
```

### Testing Error Responses

```go
assert.Error(t, err)
assert.Contains(t, err.Error(), "expected error message")

// For AppError
appErr, ok := err.(*apperror.AppError)
require.True(t, ok)
assert.Equal(t, 404, appErr.StatusCode)
assert.Equal(t, "NOT_FOUND", appErr.Code)
```

## Troubleshooting

### Issue: Tests hang or timeout
**Solution**: Check for goroutines that don't finish, use context with timeout

### Issue: Mock expectations not met
**Solution**: Verify mock setup matches actual calls, check parameter matchers

### Issue: Race conditions in tests
**Solution**: Run with race detector: `go test -race ./...`

### Issue: Tests pass locally but fail in CI
**Solution**: Check for environment-specific dependencies, timing issues

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [GORM Testing Guide](https://gorm.io/docs/testing.html)

## Contributing

When adding new features:
1. Write tests first (TDD approach)
2. Ensure tests pass: `make test`
3. Check coverage: `make test-coverage`
4. Run benchmarks if performance-critical: `make test-benchmark`

## Test Maintenance

- Review and update tests when changing implementations
- Remove obsolete tests
- Keep mocks in sync with interfaces
- Update this documentation as needed
