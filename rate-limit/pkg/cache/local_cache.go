package cache

import (
	"log"
	"time"

	"github.com/patrickmn/go-cache"
)

// LocalCache 使用patrickmn/go-cache库实现的本地缓存
type LocalCache struct {
	cache *cache.Cache
}

// NewLocalCache 创建一个新的本地缓存实例
// defaultExpiration: 默认的过期时间
// cleanupInterval: 清理过期项的时间间隔
func NewLocalCache(defaultExpiration, cleanupInterval time.Duration) *LocalCache {
	c := cache.New(defaultExpiration, cleanupInterval)
	log.Printf("Local cache initialized with default expiration: %v", defaultExpiration)
	return &LocalCache{
		cache: c,
	}
}

// Get 获取缓存中的值
func (lc *LocalCache) Get(key string) (string, bool) {
	if value, found := lc.cache.Get(key); found {
		return value.(string), true
	}
	return "", false
}

// Set 设置缓存值，带过期时间
func (lc *LocalCache) Set(key string, value string, duration time.Duration) {
	lc.cache.Set(key, value, duration)
}

// Delete 删除缓存项
func (lc *LocalCache) Delete(key string) {
	lc.cache.Delete(key)
	log.Printf("Cache item deleted: %s", key)
}

// Count 返回缓存中的条目数量
func (lc *LocalCache) Count() int {
	return lc.cache.ItemCount()
}

// Flush 清空所有缓存
func (lc *LocalCache) Flush() {
	lc.cache.Flush()
	log.Printf("Cache flushed")
}

// GetMultiple 批量获取多个key的值
func (lc *LocalCache) GetMultiple(keys []string) map[string]string {
	result := make(map[string]string)
	for _, key := range keys {
		if val, found := lc.Get(key); found {
			result[key] = val
		}
	}
	return result
}
