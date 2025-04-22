package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"multi-level-cache/internal/config"
	"multi-level-cache/pkg/utils"
)

// RedisCache 实现基于Redis的缓存
type RedisCache struct {
	name              string
	client            *redis.Client
	defaultExpiration time.Duration
}

// NewRedisCache 创建一个新的Redis缓存实例
func NewRedisCache(cfg *config.RedisConfig, opts ...Options) (*RedisCache, error) {
	options := Options{
		Name:              "redis_cache",
		DefaultExpiration: 5 * time.Minute,
	}
	if len(opts) > 0 {
		options = opts[0]
	}
	if cfg == nil {
		return nil, ErrCacheInternal
	}
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
	utils.LogInfo("Redis cache initialized: %s at %s", options.Name, cfg.Addr)
	return &RedisCache{
		name:              options.Name,
		client:            client,
		defaultExpiration: options.DefaultExpiration,
	}, nil
}

// Get 从Redis获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}
	val, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		utils.LogError("Redis GET error: %v", err)
		return nil, ErrCacheInternal
	}
	return val, nil
}

// Set 设置Redis缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	if value == nil {
		return ErrInvalidValue
	}
	if expiration <= 0 {
		expiration = r.defaultExpiration
	}
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		utils.LogError("Redis SET error: %v", err)
		return ErrCacheInternal
	}
	return nil
}

// Delete 删除Redis缓存
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		utils.LogError("Redis DEL error: %v", err)
		return ErrCacheInternal
	}
	return nil
}

// Exists 检查Redis中是否存在指定key
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		utils.LogError("Redis EXISTS error: %v", err)
		return false, ErrCacheInternal
	}
	return res > 0, nil
}

// Name 返回缓存名称
func (r *RedisCache) Name() string {
	return r.name
}

// Close 关闭Redis连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}
