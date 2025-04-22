package metrics

import (
	"fmt"
	"sync"
	"time"
)

// CacheMetrics 用于统计缓存命中、未命中等指标
type CacheMetrics struct {
	mu        sync.RWMutex
	hitCount  int64 // 命中次数
	missCount int64 // 未命中次数
	setCount  int64 // set操作次数
	delCount  int64 // delete操作次数
}

// NewCacheMetrics 创建新的指标统计实例
func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{}
}

// IncHit 命中次数加一
func (m *CacheMetrics) IncHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hitCount++
}

// IncMiss 未命中次数加一
func (m *CacheMetrics) IncMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.missCount++
}

// IncSet set操作次数加一
func (m *CacheMetrics) IncSet() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setCount++
}

// IncDel delete操作次数加一
func (m *CacheMetrics) IncDel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delCount++
}

// Snapshot 返回当前指标快照
func (m *CacheMetrics) Snapshot() (hit, miss, set, del int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hitCount, m.missCount, m.setCount, m.delCount
}

// PrintMetrics 打印当前指标
func (m *CacheMetrics) PrintMetrics() {
	hit, miss, set, del := m.Snapshot()
	fmt.Printf("[METRICS] %s | hit: %d | miss: %d | set: %d | del: %d\n",
		time.Now().Format(time.RFC3339), hit, miss, set, del)
}
