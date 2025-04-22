# Redis 热点Key限流降级系统

## 系统概述

这是一个用于处理Redis热点Key的限流降级系统，能够自动检测热点Key并采取措施减轻Redis服务器的压力。

主要特性：
- **热点Key检测**：使用滑动窗口计数器识别访问频率高的Key
- **访问限流**：对热点Key应用令牌桶算法进行限流
- **本地缓存**：将热点Key缓存到本地内存，减少对Redis的请求
- **自适应降级**：当访问量超过阈值时自动降级到本地缓存

通过这种多层防护机制，系统能够在高并发场景下有效保护Redis服务器，防止单个热点Key导致的性能问题。

## 系统架构

```
┌─────────────────┐         ┌─────────────────┐
│                 │         │                 │
│    客户端请求    │─────────▶     API 服务     │
│                 │         │                 │
└─────────────────┘         └────────┬────────┘
                                    │
                                    ▼
           ┌─────────────────────────────────────┐
           │         热点Key限流降级系统           │
           │                                     │
           │  ┌─────────────┐   ┌─────────────┐  │
           │  │  热点检测器   │◀──▶│   限流器    │  │
           │  └─────────────┘   └─────────────┘  │
           │         │                │          │
           │         ▼                ▼          │
           │  ┌─────────────┐   ┌─────────────┐  │
           │  │  本地缓存    │◀──▶│ Redis客户端 │  │
           │  └─────────────┘   └─────────────┘  │
           │                         │           │
           └─────────────────────────┼───────────┘
                                    │
                                    ▼
                          ┌─────────────────┐
                          │                 │
                          │  Redis 服务器   │
                          │                 │
                          └─────────────────┘
```

## 核心组件

### 1. 热点Key检测器 (HotKeyDetector)

负责识别和标记热点Key：

- **访问计数**：记录每个Key在时间窗口内的访问次数
- **阈值判断**：当访问次数超过预设阈值时将Key标记为热点
- **标记过期**：热点Key标记具有自动过期功能，适应流量变化

```go
// 记录Key访问并检查是否为热点
func (d *HotKeyDetector) RecordAccess(key string) bool {
    // 检查Key是否已经是热点
    if _, isHot := d.hotKeys.Get(key); isHot {
        return true
    }
    
    // 更新访问计数
    count := updateCounter(key)
    
    // 检查是否超过阈值
    if count >= d.config.Threshold {
        log.Printf("Hot key detected: %s with %d accesses in %v", 
                   key, count, d.config.Window)
        d.hotKeys.Set(key, "true", d.config.HotKeyExpiration)
        return true
    }
    
    return false
}
```

### 2. 限流器 (RateLimiter)

基于令牌桶算法的限流组件：

- **令牌桶**：为每个热点Key维护独立的令牌桶
- **限流控制**：控制热点Key的访问频率，防止过载
- **自动清理**：定期清理不再活跃的限流器实例

```go
// 检查是否允许访问
func (rl *RateLimiter) Allow(key string) bool {
    limiter := rl.getLimiter(key)
    allowed := limiter.Allow()
    if !allowed {
        log.Printf("Rate limited: %s", key)
    }
    return allowed
}
```

### 3. 本地缓存 (LocalCache)

提供高效的内存缓存服务：

- **快速访问**：为热点Key提供内存级别的访问速度
- **自动过期**：设置合理的过期时间确保数据一致性
- **容量控制**：避免内存过度使用

### 4. Redis客户端 (RedisClient)

封装与Redis服务器的交互：

- **连接管理**：维护与Redis的连接
- **键值操作**：提供存取、删除等基本操作
- **错误处理**：统一处理Redis操作中的异常

## 热点Key处理原理

### 1. 检测阶段

系统使用滑动窗口计数器跟踪Key的访问频率：

1. **记录访问**：每次Key被访问时增加计数器
2. **窗口限定**：只统计固定时间窗口内的访问次数
3. **标记热点**：当访问次数超过阈值时标记为热点Key

### 2. 限流阶段

对识别出的热点Key实施限流策略：

1. **创建限流器**：为每个热点Key创建专用限流器
2. **令牌分配**：按配置速率为限流器补充令牌
3. **访问控制**：请求到达时消耗令牌，无令牌时拒绝访问

### 3. 降级阶段

当热点Key被限流时系统会自动降级：

1. **本地缓存**：热点Key的值被缓存在本地内存中
2. **缓存优先**：优先从本地缓存获取热点Key的值
3. **缓存更新**：在允许的情况下从Redis更新本地缓存

## 如何运行系统

### 前提条件

- Go 1.16+
- Redis 服务器实例

### 配置

系统默认使用本地Redis实例：

- 地址: `localhost:6379`
- 无密码
- 数据库: 0

如需修改，可以调整`pkg/storage/redis_client.go`中的`DefaultConfig`。

### 启动步骤

1. **安装依赖**：
   ```bash
   go mod tidy
   ```

2. **运行程序**：
   ```bash
   go run cmd/main.go
   ```

3. **观察日志**：
   程序会输出启动信息和热点Key检测日志。

### 测试系统功能

1. **设置一个键值**：
   ```bash
   curl -X POST "http://localhost:8080/set/testkey" -d "value=hello"
   ```

2. **模拟大量访问制造热点Key**：
   ```bash
   # Windows PowerShell
   for ($i=0; $i -lt 200; $i++) { 
       Invoke-WebRequest -Uri "http://localhost:8080/get/testkey" -Method GET > $null
   }

   # Linux/Mac
   for i in {1..200}; do 
       curl -s "http://localhost:8080/get/testkey" > /dev/null
   done
   ```

3. **查看Key统计信息**：
   ```bash
   curl "http://localhost:8080/stats/testkey"
   ```

4. **测试是否触发限流**：
   ```bash
   curl -v "http://localhost:8080/get/testkey"
   ```

## 代码结构

- `cmd/`: 应用入口
    - main.go: 主程序，初始化并启动服务

- `pkg/`: 核心组件包
    - `detector/`: 热点Key检测
        - hotkey_detector.go: 热点Key检测器实现
    - `limiter/`: 限流功能
        - rate_limiter.go: 令牌桶限流器
    - `cache/`: 缓存相关
        - local_cache.go: 本地内存缓存
    - `storage/`: 存储相关
        - redis_client.go: Redis客户端封装

- `api/`: API服务
    - server.go: HTTP服务器和路由处理

## 主要工作流程

1. **客户端请求Key**：`GET /get/{key}`
2. **系统检测是否为热点Key**
3. **若为热点Key**：
    - 尝试从本地缓存获取
    - 如果缓存未命中，检查限流器是否允许访问Redis
    - 若不允许，返回限流错误(429状态码)
    - 若允许，从Redis获取并更新本地缓存
4. **若非热点Key**：
    - 直接从Redis获取
    - 更新访问计数