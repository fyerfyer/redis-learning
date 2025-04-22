package cache

import (
	"context"
	"errors"
	"time"
)

// 常见错误定义
var (
	ErrKeyNotFound   = errors.New("key not found in cache")
	ErrInvalidKey    = errors.New("invalid key")
	ErrInvalidValue  = errors.New("invalid value")
	ErrCacheInternal = errors.New("internal cache error")
)

// Cache 定义缓存的基本操作接口
type Cache interface {
	// Get 获取缓存的值，如果不存在则返回 ErrKeyNotFound
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存的值，可选过期时间
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete 删除缓存的值
	Delete(ctx context.Context, key string) error

	// Exists 检查key是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Name 返回缓存实现的名称，用于日志和指标
	Name() string

	// Close 关闭并清理缓存资源
	Close() error
}

// Options 定义缓存的配置选项
type Options struct {
	// 缓存的名称
	Name string

	// 默认过期时间，如果为0则表示不过期
	DefaultExpiration time.Duration
}
