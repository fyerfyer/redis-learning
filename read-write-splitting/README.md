# Redis 读写分离系统

一个简单的 Redis 读写分离代理系统，用于减轻 Redis 主库的访问压力并提高系统的读取性能。

## 系统架构

```
┌─────────────────────────┐
│                         │
│     应用程序/客户端       │
│                         │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│     读写分离代理系统      │
├─────────────────────────┤
│     RedisProxy          │
├─────────────────────────┤
│     负载均衡器           │
└───┬─────────────────┬───┘
    │                 │
    ▼                 ▼
┌─────────┐     ┌─────────────┐
│ Redis   │     │ Redis Slaves │
│ Master  │     │ (多个从库)   │
└─────────┘     └─────────────┘
```

## 核心组件

### 1. 读写分离代理 (RedisProxy)

处理所有 Redis 命令并根据命令类型路由到适当的实例：

- **写命令路由**：将写操作定向到主库
- **读命令路由**：将读操作分发到从库
- **故障转移**：从库不可用时自动降级到主库
- **健康检查**：定期检查从库健康状态

```go
// Process 处理Redis命令
func (p *RedisProxy) Process(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
    if p.IsWriteCommand(cmd) {
        // 写命令路由到主库
        return p.processOnMaster(ctx, cmd, args...)
    } else {
        // 读命令路由到从库
        return p.processOnSlave(ctx, cmd, args...)
    }
}
```

### 2. 负载均衡器 (Balancer)

管理多个从库之间的请求分配：

- **轮询分配**：使用 Round Robin 策略分配读请求
- **健康状态管理**：跟踪从库的可用状态
- **动态调整**：自动排除不可用的从库

```go
// Next 获取下一个可用的Redis从库索引
func (b *RoundRobinBalancer) Next(slaveCount int) int {
    // 增加计数器并获取当前值
    current := atomic.AddUint64(&b.counter, 1) - 1
    
    // 尝试找到一个可用的从库
    for i := 0; i < slaveCount; i++ {
        index := int((current + uint64(i)) % uint64(slaveCount))
        if b.isSlaveAvailable(index) {
            return index
        }
    }
    
    // 所有从库都不可用
    return -1
}
```

### 3. 配置管理 (Config)

管理 Redis 主从集群的连接信息：

- **主从配置**：维护主库和多个从库的连接信息
- **连接池设置**：管理 Redis 连接池的大小与配置
- **默认配置**：提供开箱即用的默认设置

## 读写分离原理

### 1. 命令分类

系统基于命令类型进行路由：

- **写命令**：如 SET, DEL, INCR, LPUSH 等修改数据的操作路由到主库
- **读命令**：如 GET, HGETALL, SMEMBERS 等只读操作路由到从库
- **未知命令**：默认路由到主库以确保数据安全

### 2. 读操作流程

读请求的处理流程：

1. 系统识别命令为读操作
2. 负载均衡器选择一个可用的从库
3. 将请求发送到选中的从库
4. 如果从库操作失败，尝试其他从库
5. 如果所有从库都不可用，降级到主库

### 3. 写操作流程

写请求的处理流程：

1. 系统识别命令为写操作
2. 请求直接路由到主库执行
3. 写操作的结果返回给客户端
4. 主库的变更会通过 Redis 的主从复制机制同步到从库

### 4. 容错机制

系统提供多层容错保护：

- **从库故障**：如果某个从库不可用，负载均衡器会自动排除该实例
- **全部从库故障**：在所有从库不可用时，系统会自动降级到主库
- **健康检查**：定期检查并恢复已修复的从库

## 如何运行系统

### 前提条件

- Go 1.18 或更高版本
- Redis 服务器（至少一个主库和一个从库）

### 配置系统

1. **Redis 主从配置**：

   确保已设置 Redis 主从复制：

   ```bash
   # 从库配置示例
   redis-server --port 6380 --slaveof 127.0.0.1 6379
   redis-server --port 6381 --slaveof 127.0.0.1 6379
   ```

2. **更新代理配置**：

   修改 config.go 中的配置以匹配 Redis 部署：

   ```go
   func DefaultConfig() *RedisClusterConfig {
       return &RedisClusterConfig{
           Master: RedisConfig{
               Host:     "localhost",
               Port:     6379,    // 主库端口
               Password: "",      // 如需密码请设置
               DB:       0,
           },
           Slaves: []RedisConfig{
               {
                   Host:     "localhost",
                   Port:     6380,  // 从库1
                   Password: "",
                   DB:       0,
               },
               {
                   Host:     "localhost",
                   Port:     6381,  // 从库2
                   Password: "",
                   DB:       0,
               },
           },
           PoolSize: 10,
       }
   }
   ```

### 启动步骤

1. **构建并运行程序**：
   ```bash
   cd read-write-splitting
   go build -o rw-proxy cmd/main.go
   ./rw-proxy
   ```

   或者直接运行：
   ```bash
   cd read-write-splitting
   go run cmd/main.go
   ```

## 示例用法

```go
package main

import (
    "context"
    "fmt"
    "time"

    "read-write-splitting/internal/config"
    "read-write-splitting/proxy"
)

func main() {
    // 获取默认配置
    cfg := config.DefaultConfig()

    // 初始化Redis读写分离代理
    rp := proxy.NewRedisProxy(cfg)
    defer rp.Close()

    // 启动健康检查
    rp.StartHealthCheck(10 * time.Second)

    // 使用上下文
    ctx := context.Background()

    // 写入数据 (路由到主库)
    _, err := rp.Process(ctx, "set", "user:1", "张三")
    if err != nil {
        fmt.Printf("写入失败: %v\n", err)
    }

    // 读取数据 (路由到从库)
    value, err := rp.Process(ctx, "get", "user:1")
    if err != nil {
        fmt.Printf("读取失败: %v\n", err)
    } else {
        fmt.Printf("用户数据: %v\n", value)
    }
}
```

## 代码结构

- `cmd/`: 命令行入口
    - main.go: 主程序，展示系统用例

- `internal/`: 内部实现
    - `config/`: 配置管理
        - config.go: Redis 主从配置

- `proxy/`: 代理实现
    - proxy.go: 读写分离核心逻辑
    - balancer.go: 负载均衡器实现

- `go.mod`: Go 模块定义文件
