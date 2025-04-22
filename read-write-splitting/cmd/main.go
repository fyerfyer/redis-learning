package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"read-write-splitting/internal/config"
	"read-write-splitting/proxy"
)

func main() {
	// 创建默认配置
	cfg := &config.RedisClusterConfig{
		Master: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Slaves: []config.RedisConfig{
			{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
		},
		PoolSize: 10,
	}

	// 初始化Redis读写分离代理
	redisProxy := proxy.NewRedisProxy(cfg)
	defer redisProxy.Close()

	// 启动健康检查
	redisProxy.StartHealthCheck(10 * time.Second)

	// 创建上下文
	ctx := context.Background()

	// 演示写操作 - 将路由到主库
	fmt.Println("===== Write Operation Examples =====")
	_, err := redisProxy.Process(ctx, "set", "user:1", "John Doe")
	if err != nil {
		fmt.Printf("Failed to set key: %v\n", err)
	} else {
		fmt.Println("Successfully set key 'user:1'")
	}

	_, err = redisProxy.Process(ctx, "set", "counter", "1")
	if err != nil {
		fmt.Printf("Failed to set counter: %v\n", err)
	} else {
		fmt.Println("Successfully set key 'counter'")
	}

	// 演示读操作 - 将路由到从库
	fmt.Println("\n===== Read Operation Examples =====")
	val, err := redisProxy.Process(ctx, "get", "user:1")
	if err != nil {
		fmt.Printf("Failed to get key: %v\n", err)
	} else {
		fmt.Printf("Value for 'user:1': %v\n", val)
	}

	// 演示计数器递增 - 写操作，路由到主库
	fmt.Println("\n===== Increment Operation Examples =====")
	_, err = redisProxy.Process(ctx, "incr", "counter")
	if err != nil {
		fmt.Printf("Failed to increment counter: %v\n", err)
	} else {
		fmt.Println("Successfully incremented 'counter'")
	}

	// 读取递增后的计数器值 - 读操作，路由到从库
	val, err = redisProxy.Process(ctx, "get", "counter")
	if err != nil {
		fmt.Printf("Failed to get counter: %v\n", err)
	} else {
		fmt.Printf("Value for 'counter': %v\n", val)
	}

	// 演示哈希表操作
	fmt.Println("\n===== Hash Operation Examples =====")
	_, err = redisProxy.Process(ctx, "hset", "user:profile:1", "name", "John", "age", "30", "city", "New York")
	if err != nil {
		fmt.Printf("Failed to set hash: %v\n", err)
	} else {
		fmt.Println("Successfully set hash 'user:profile:1'")
	}

	val, err = redisProxy.Process(ctx, "hget", "user:profile:1", "name")
	if err != nil {
		fmt.Printf("Failed to get hash field: %v\n", err)
	} else {
		fmt.Printf("Name from hash: %v\n", val)
	}

	val, err = redisProxy.Process(ctx, "hgetall", "user:profile:1")
	if err != nil {
		fmt.Printf("Failed to get entire hash: %v\n", err)
	} else {
		fmt.Printf("All hash fields: %v\n", val)
	}

	// 优雅退出
	fmt.Println("\nPress Ctrl+C to exit...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down Redis connections...")
}
