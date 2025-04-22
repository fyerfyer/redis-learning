package limiter

import (
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	// 每秒允许的请求数
	RatePerSecond float64
	// 桶容量（允许的突发请求数）
	BurstSize int
}

// DefaultRateLimiterConfig 默认限流配置
var DefaultRateLimiterConfig = RateLimiterConfig{
	RatePerSecond: 10.0, // 每秒10个请求
	BurstSize:     20,   // 允许20个突发请求
}

// RateLimiter 基于令牌桶算法的限流器
type RateLimiter struct {
	config       RateLimiterConfig
	limiters     map[string]*rate.Limiter
	limiterMutex sync.RWMutex
	cleanupTime  time.Duration
}

// NewRateLimiter 创建一个新的限流器
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:       config,
		limiters:     make(map[string]*rate.Limiter),
		limiterMutex: sync.RWMutex{},
		cleanupTime:  time.Hour, // 默认1小时清理一次不再使用的限流器
	}

	// 启动一个协程定期清理不再使用的限流器
	go rl.cleanup()

	return rl
}

// NewDefaultRateLimiter 使用默认配置创建限流器
func NewDefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(DefaultRateLimiterConfig)
}

// Allow 检查指定key的访问是否被允许
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.getLimiter(key)
	allowed := limiter.Allow()
	if !allowed {
		log.Printf("Rate limited: %s", key)
	}
	return allowed
}

// getLimiter 获取指定key的限流器，如果不存在则创建
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.limiterMutex.RLock()
	limiter, exists := rl.limiters[key]
	rl.limiterMutex.RUnlock()

	if exists {
		return limiter
	}

	// 如果不存在，创建一个新的限流器
	rl.limiterMutex.Lock()
	defer rl.limiterMutex.Unlock()

	// 再次检查，可能在获取写锁的过程中已经被其他协程创建
	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}

	// 创建一个新的限流器
	limiter = rate.NewLimiter(rate.Limit(rl.config.RatePerSecond), rl.config.BurstSize)
	rl.limiters[key] = limiter
	log.Printf("Created new rate limiter for: %s", key)

	return limiter
}

// cleanup 定期清理不再使用的限流器
// 这是一个简化版，实际上我们可能需要记录最后使用时间来决定是否清理
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupTime)
	defer ticker.Stop()

	for range ticker.C {
		rl.limiterMutex.Lock()
		// 简单实现，实际生产环境可能需要更复杂的清理逻辑
		count := len(rl.limiters)
		// 这里简单粗暴地定期重置所有限流器
		// 实际应用中可能需要更精细的策略
		rl.limiters = make(map[string]*rate.Limiter)
		rl.limiterMutex.Unlock()

		log.Printf("Cleaned up %d rate limiters", count)
	}
}

// SetRateForKey 为特定key设置自定义限流速率
func (rl *RateLimiter) SetRateForKey(key string, ratePerSecond float64, burstSize int) {
	rl.limiterMutex.Lock()
	defer rl.limiterMutex.Unlock()

	// 创建或更新限流器
	rl.limiters[key] = rate.NewLimiter(rate.Limit(ratePerSecond), burstSize)
	log.Printf("Set custom rate for %s: %.2f req/s, burst: %d", key, ratePerSecond, burstSize)
}
