package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) SAdd(ctx context.Context, key string, members ...any) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepository) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) SRem(ctx context.Context, key string, members ...any) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheRepository) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	args := m.Called(ctx, key, member)
	return args.Bool(0), args.Error(1)
}