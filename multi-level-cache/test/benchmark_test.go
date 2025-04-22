package test

import (
	"context"
	"testing"
	"time"

	"multi-level-cache/internal/cache"
	"multi-level-cache/internal/config"
)

// BenchmarkMultiLevelCache_Get 测试多级缓存Get性能
func BenchmarkMultiLevelCache_Get(b *testing.B) {
	ctx := context.Background()
	local, _ := cache.NewLocalCache(&config.DefaultConfig().LocalCache)
	redis, _ := cache.NewRedisCache(&config.DefaultConfig().Redis)
	mc := cache.NewMultiLevelCache(local, redis)

	key := "bench_key"
	value := []byte("bench_value")
	_ = mc.Set(ctx, key, value, 1*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mc.Get(ctx, key)
	}
}

// BenchmarkMultiLevelCache_Set 测试多级缓存Set性能
func BenchmarkMultiLevelCache_Set(b *testing.B) {
	ctx := context.Background()
	local, _ := cache.NewLocalCache(&config.DefaultConfig().LocalCache)
	redis, _ := cache.NewRedisCache(&config.DefaultConfig().Redis)
	mc := cache.NewMultiLevelCache(local, redis)

	key := "bench_key"
	value := []byte("bench_value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mc.Set(ctx, key, value, 1*time.Minute)
	}
}

// BenchmarkMultiLevelCache_ParallelGet 并发测试多级缓存Get
func BenchmarkMultiLevelCache_ParallelGet(b *testing.B) {
	ctx := context.Background()
	local, _ := cache.NewLocalCache(&config.DefaultConfig().LocalCache)
	redis, _ := cache.NewRedisCache(&config.DefaultConfig().Redis)
	mc := cache.NewMultiLevelCache(local, redis)

	key := "bench_key"
	value := []byte("bench_value")
	_ = mc.Set(ctx, key, value, 1*time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = mc.Get(ctx, key)
		}
	})
}
