package cache

import (
	"context"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"multi-level-cache/internal/config"
	"multi-level-cache/pkg/utils"
)

// LocalCache 实现基于内存的本地缓存
type LocalCache struct {
	// 缓存名称
	name string
	// 内部缓存实例
	cache *cache.Cache
	// 默认过期时间
	defaultExpiration time.Duration
	// 互斥锁，用于一些需要同步的操作
	mu sync.RWMutex
}

// NewLocalCache 创建一个新的本地缓存
func NewLocalCache(cfg *config.LocalCacheConfig, opts ...Options) (*LocalCache, error) {
	options := Options{
		Name:              "local_cache",
		DefaultExpiration: 5 * time.Minute,
	}

	// 应用自定义选项
	if len(opts) > 0 {
		options = opts[0]
	}

	// 如果提供了配置，则使用配置的值
	if cfg != nil {
		if cfg.DefaultExpiration > 0 {
			options.DefaultExpiration = cfg.DefaultExpiration
		}
	}

	c := &LocalCache{
		name:              options.Name,
		cache:             cache.New(options.DefaultExpiration, options.DefaultExpiration*2),
		defaultExpiration: options.DefaultExpiration,
	}

	utils.LogInfo("Local cache initialized: %s with default expiration: %v", options.Name, options.DefaultExpiration)
	return c, nil
}

// Get 从本地缓存获取值
func (c *LocalCache) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 从缓存获取值
	value, found := c.cache.Get(key)
	if !found {
		return nil, ErrKeyNotFound
	}

	// 将值转换为字节数组
	bytes, ok := value.([]byte)
	if !ok {
		utils.LogError("Invalid type in cache for key: %s", key)
		return nil, ErrCacheInternal
	}

	return bytes, nil
}

// Set 设置缓存值
func (c *LocalCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if value == nil {
		return ErrInvalidValue
	}

	// 如果未指定过期时间，则使用默认过期时间
	if expiration <= 0 {
		expiration = c.defaultExpiration
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Set(key, value, expiration)
	return nil
}

// Delete 从缓存中删除键
func (c *LocalCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Delete(key)
	return nil
}

// Exists 检查键是否存在于缓存中
func (c *LocalCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.cache.Get(key)
	return found, nil
}

// Name 返回缓存名称
func (c *LocalCache) Name() string {
	return c.name
}

// Close 清理缓存资源
func (c *LocalCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空缓存
	c.cache.Flush()
	return nil
}
