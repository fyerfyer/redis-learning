# Multi-Level Cache

一个简单的多级缓存系统，用于减少Redis访问次数并解决热点key问题。

## 系统概述

- **一级缓存**：基于内存的本地缓存，访问速度快但容量受限
- **二级缓存**：基于Redis的共享缓存，容量大但网络访问较慢

通过多级缓存架构，系统能够有效降低对Redis的访问频率，提高热点数据的访问速度，同时保持数据的一致性。

## 项目结构

```
multi-level-cache/
├── cmd/
│   └── main.go                  # 演示程序入口
├── internal/
│   ├── cache/
│   │   ├── cache.go             # 缓存接口定义
│   │   ├── local_cache.go       # 本地内存缓存实现
│   │   ├── redis_cache.go       # Redis缓存实现
│   │   └── multi-level-cache.go # 多级缓存协调器
│   └── config/
│       └── config.go            # 配置相关
├── pkg/
│   ├── metrics/
│   │   └── metrics.go           # 简单指标收集
│   └── utils/
│       └── utils.go             # 通用工具函数
└── test/
    └── benchmark_test.go        # 性能测试
```

## 快速开始

### 前置条件

- Go 1.16+
- Redis 服务器

### 安装

```bash
git clone https://github.com/yourusername/multi-level-cache.git
cd multi-level-cache
go mod tidy
```

### 运行示例

```bash
go run cmd/main.go
```

## 使用示例

```go
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
	local, _ := cache.NewLocalCache(&cfg.LocalCache)
	redis, _ := cache.NewRedisCache(&cfg.Redis)

	// 创建多级缓存
	mc := cache.NewMultiLevelCache(local, redis)

	// 使用缓存
	ctx := context.Background()
	mc.Set(ctx, "user:1001", []byte(`{"name":"张三","age":30}`), 5*time.Minute)
	
	// 获取缓存数据
	val, _ := mc.Get(ctx, "user:1001") 
	fmt.Printf("用户数据: %s\n", string(val))
	
	// 打印缓存指标
	mc.PrintMetrics()
}
```

## 配置说明

在 config.go 中可以自定义以下配置：

### Redis配置

```go
RedisConfig{
    Addr:         "localhost:6379",  // Redis服务器地址
    Password:     "",                // Redis密码
    DB:           0,                 // 数据库索引
    PoolSize:     10,                // 连接池大小
    DialTimeout:  5 * time.Second,   // 连接超时
    ReadTimeout:  3 * time.Second,   // 读取超时
    WriteTimeout: 3 * time.Second,   // 写入超时
}
```

### 本地缓存配置

```go
LocalCacheConfig{
    MaxEntries:        1000,               // 最大条目数
    DefaultExpiration: 5 * time.Minute,    // 默认过期时间
    CleanupInterval:   10 * time.Minute,   // 清理间隔
}
```

## 性能测试

运行基准测试来评估缓存性能：

```bash
cd test
go test -bench=.
```

## 多级缓存原理

1. **读取流程**：
    - 首先尝试从本地缓存读取数据
    - 如果本地缓存未命中，则从Redis读取
    - 从Redis读取成功后，自动回写到本地缓存

2. **写入流程**：
    - 同时写入本地缓存和Redis
    - 本地缓存可设置较短的过期时间

3. **删除流程**：
    - 同时删除本地缓存和Redis中的数据