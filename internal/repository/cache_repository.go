package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SAdd(ctx context.Context, key string, members ...any) error
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Del(ctx context.Context, key string) error
	SRem(ctx context.Context, key string, members ...any) error
	SIsMember(ctx context.Context, key string, member any) (bool, error)
}

type cacheRepository struct {
	rdb *redis.UniversalClient
}

func NewCacheRepository(rdb *redis.UniversalClient) CacheRepository {
	return &cacheRepository{rdb: rdb}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
    return (*r.rdb).Set(ctx, key, value, expiration).Err()
}

func (r *cacheRepository) SAdd(ctx context.Context, key string, members ...any) error {
    return (*r.rdb).SAdd(ctx, key, members...).Err()
}

func (r *cacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
    return (*r.rdb).Expire(ctx, key, expiration).Result()
}

func (r *cacheRepository) Del(ctx context.Context, key string) error {
    return (*r.rdb).Del(ctx, key).Err()
}

func (r *cacheRepository) SRem(ctx context.Context, key string, members ...any) error {
    return (*r.rdb).SRem(ctx, key, members...).Err()
}

func (r *cacheRepository) SIsMember(ctx context.Context, key string, member any) (bool, error) {
    return (*r.rdb).SIsMember(ctx, key, member).Result()
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
    return (*r.rdb).Get(ctx, key).Result()
}