package detector

import (
	"log"
	"sync"
	"time"

	"rate-limit/pkg/cache"
)

// HotKeyConfig 热点key检测器配置
type HotKeyConfig struct {
	// 访问阈值，超过此值将被视为热点key
	Threshold int64
	// 统计窗口，在此时间范围内统计访问次数
	Window time.Duration
	// 热点key的过期时间
	HotKeyExpiration time.Duration
}

// DefaultHotKeyConfig 默认热点key检测配置
var DefaultHotKeyConfig = HotKeyConfig{
	Threshold:        100,              // 100次访问视为热点
	Window:           time.Second * 10, // 10秒内
	HotKeyExpiration: time.Minute * 5,  // 热点key标记5分钟后过期
}

// HotKeyDetector 热点key检测器
type HotKeyDetector struct {
	config      HotKeyConfig
	localCache  *cache.LocalCache
	counterLock sync.RWMutex
	hotKeys     *cache.LocalCache // 用于存储热点key
}

// NewHotKeyDetector 创建一个新的热点key检测器
func NewHotKeyDetector(config HotKeyConfig) *HotKeyDetector {
	// 创建两个缓存：一个用于计数，一个用于存储热点key
	counterCache := cache.NewLocalCache(config.Window, time.Minute)
	hotKeysCache := cache.NewLocalCache(config.HotKeyExpiration, time.Minute)

	return &HotKeyDetector{
		config:      config,
		localCache:  counterCache,
		counterLock: sync.RWMutex{},
		hotKeys:     hotKeysCache,
	}
}

// NewDefaultHotKeyDetector 使用默认配置创建热点key检测器
func NewDefaultHotKeyDetector() *HotKeyDetector {
	return NewHotKeyDetector(DefaultHotKeyConfig)
}

// RecordAccess 记录key的访问并检测是否为热点key
func (d *HotKeyDetector) RecordAccess(key string) bool {
	// 检查key是否已经是热点key
	if _, isHot := d.hotKeys.Get(key); isHot {
		return true
	}

	// 更新访问计数
	d.counterLock.Lock()
	defer d.counterLock.Unlock()

	count := int64(1)
	if val, exists := d.localCache.Get(key); exists {
		currentCount, _ := time.ParseDuration(val)
		count = int64(currentCount) + 1
	}

	// 将count转换为string存储
	d.localCache.Set(key, (time.Duration(count)).String(), d.config.Window)

	// 检查是否超过阈值
	if count >= d.config.Threshold {
		log.Printf("Hot key detected: %s with %d accesses in %v", key, count, d.config.Window)
		d.hotKeys.Set(key, "true", d.config.HotKeyExpiration)
		return true
	}

	return false
}

// IsHotKey 检查key是否是热点key
func (d *HotKeyDetector) IsHotKey(key string) bool {
	_, isHot := d.hotKeys.Get(key)
	return isHot
}

// GetAccessCount 获取key的访问次数
func (d *HotKeyDetector) GetAccessCount(key string) int64 {
	d.counterLock.RLock()
	defer d.counterLock.RUnlock()

	if val, exists := d.localCache.Get(key); exists {
		currentCount, _ := time.ParseDuration(val)
		return int64(currentCount)
	}
	return 0
}

// GetHotKeys 获取所有热点key
func (d *HotKeyDetector) GetHotKeys() []string {
	// 这里实现一个简单版本，实际上go-cache没有提供直接获取所有键的方法
	// 在实际应用中，我们可能需要另外维护一个热点key的列表
	return []string{}
}

// ClearHotKey 清除指定key的热点标记
func (d *HotKeyDetector) ClearHotKey(key string) {
	d.hotKeys.Delete(key)
	log.Printf("Hot key mark removed: %s", key)
}
