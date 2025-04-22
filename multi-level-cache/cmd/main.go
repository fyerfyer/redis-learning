package main

import (
	"context"
	"fmt"
	"time"

	"multi-level-cache/internal/cache"
	"multi-level-cache/internal/config"
)

func main() {
	// 初始化配置
	cfg := config.DefaultConfig()

	// 创建本地缓存和Redis缓存
	local, err := cache.NewLocalCache(&cfg.LocalCache)
	if err != nil {
		fmt.Printf("Failed to init local cache: %v\n", err)
		return
	}
	redis, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		fmt.Printf("Failed to init redis cache: %v\n", err)
		return
	}

	// 创建多级缓存
	mc := cache.NewMultiLevelCache(local, redis)

	ctx := context.Background()
	key := "demo_key"
	value := []byte("hello multi-level cache")

	// 写入缓存
	if err := mc.Set(ctx, key, value, 30*time.Second); err != nil {
		fmt.Printf("Set error: %v\n", err)
		return
	}
	fmt.Println("Set success")

	// 读取缓存
	val, err := mc.Get(ctx, key)
	if err != nil {
		fmt.Printf("Get error: %v\n", err)
		return
	}
	fmt.Printf("Get success, value: %s\n", string(val))

	// 检查key是否存在
	exists, _ := mc.Exists(ctx, key)
	fmt.Printf("Exists: %v\n", exists)

	// 删除缓存
	if err := mc.Delete(ctx, key); err != nil {
		fmt.Printf("Delete error: %v\n", err)
		return
	}
	fmt.Println("Delete success")

	// 再次读取，应该未命中
	_, err = mc.Get(ctx, key)
	if err != nil {
		fmt.Printf("Get after delete (should miss): %v\n", err)
	}

	// 打印缓存指标
	mc.PrintMetrics()

	// 关闭缓存资源
	_ = mc.Close()
}
