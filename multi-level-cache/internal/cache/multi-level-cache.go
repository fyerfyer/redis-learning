package cache

import (
	"context"
	"errors"
	"time"

	"multi-level-cache/pkg/metrics"
	"multi-level-cache/pkg/utils"
)

// MultiLevelCache 实现简单的多级缓存（本地缓存 + Redis缓存）
type MultiLevelCache struct {
	name    string
	local   Cache // 本地缓存
	redis   Cache // Redis缓存
	metrics *metrics.CacheMetrics
}

// MultiLevelCacheOptions 多级缓存配置选项
type MultiLevelCacheOptions struct {
	Name string
}

// NewMultiLevelCache 创建多级缓存实例
func NewMultiLevelCache(local, redis Cache, opts ...MultiLevelCacheOptions) *MultiLevelCache {
	name := "multi_level_cache"
	if len(opts) > 0 && opts[0].Name != "" {
		name = opts[0].Name
	}
	utils.LogInfo("MultiLevelCache initialized: %s", name)
	return &MultiLevelCache{
		name:    name,
		local:   local,
		redis:   redis,
		metrics: metrics.NewCacheMetrics(),
	}
}

// Get 先查本地缓存，再查Redis，最后返回
func (m *MultiLevelCache) Get(ctx context.Context, key string) ([]byte, error) {
	// 先查本地缓存
	val, err := m.local.Get(ctx, key)
	if err == nil {
		m.metrics.IncHit()
		return val, nil
	}
	if !errors.Is(err, ErrKeyNotFound) {
		m.metrics.IncMiss()
		return nil, err
	}

	// 本地未命中，查Redis
	val, err = m.redis.Get(ctx, key)
	if err == nil {
		m.metrics.IncHit()
		// 回写本地缓存，过期时间可自定义，这里简单用默认
		_ = m.local.Set(ctx, key, val, 0)
		return val, nil
	}
	if errors.Is(err, ErrKeyNotFound) {
		m.metrics.IncMiss()
	}
	return nil, err
}

// Set 同时写入本地缓存和Redis
func (m *MultiLevelCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	err1 := m.local.Set(ctx, key, value, expiration)
	err2 := m.redis.Set(ctx, key, value, expiration)
	m.metrics.IncSet()
	if err1 != nil {
		utils.LogError("Local cache set error: %v", err1)
	}
	if err2 != nil {
		utils.LogError("Redis cache set error: %v", err2)
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// Delete 同时删除本地缓存和Redis
func (m *MultiLevelCache) Delete(ctx context.Context, key string) error {
	err1 := m.local.Delete(ctx, key)
	err2 := m.redis.Delete(ctx, key)
	m.metrics.IncDel()
	if err1 != nil {
		utils.LogError("Local cache delete error: %v", err1)
	}
	if err2 != nil {
		utils.LogError("Redis cache delete error: %v", err2)
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// Exists 检查本地缓存和Redis是否存在
func (m *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	ok, err := m.local.Exists(ctx, key)
	if err == nil && ok {
		return true, nil
	}
	ok, err = m.redis.Exists(ctx, key)
	return ok, err
}

// Name 返回缓存名称
func (m *MultiLevelCache) Name() string {
	return m.name
}

// Close 关闭所有缓存资源
func (m *MultiLevelCache) Close() error {
	err1 := m.local.Close()
	err2 := m.redis.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// PrintMetrics 打印缓存命中等指标
func (m *MultiLevelCache) PrintMetrics() {
	m.metrics.PrintMetrics()
}
