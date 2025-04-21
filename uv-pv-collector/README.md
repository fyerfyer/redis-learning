# Redis UV/PV 统计系统

## 系统架构

```
┌─────────────────────────┐
│                         │
│      Web应用/客户端      │
│                         │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│    UV/PV 统计系统        │
├─────────────────────────┤
│  StatsService           │      ┌─────────────┐
├─────────────────────────┤      │             │
│  StatsCollector         │◄────►│  HTTP API   │
├─────────────────────────┤      │             │
│  StatsHandler           │      └─────────────┘
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│         Redis           │
│                         │
└─────────────────────────┘
```

## 核心组件

### 1. 统计服务 (StatsService)

直接与Redis交互，提供底层的PV和UV统计功能：

- **PV统计**：使用Redis的INCR命令增加页面浏览计数
- **UV统计**：使用Redis的HyperLogLog数据结构记录唯一访客
- **数据查询**：获取指定页面和日期的统计数据

```go
// 记录页面浏览量(PV)
func (s *StatsService) RecordPageView(ctx context.Context, page string) error {
    date := time.Now().Format("2006-01-02")
    key := fmt.Sprintf("pv:%s:%s", page, date)
    
    // 使用INCR命令增加计数器
    if err := s.redisClient.Incr(ctx, key).Err(); err != nil {
        return fmt.Errorf("failed to record page view: %w", err)
    }
    
    return nil
}

// 记录唯一访客(UV)
func (s *StatsService) RecordUniqueVisitor(ctx context.Context, page, visitorID string) error {
    date := time.Now().Format("2006-01-02")
    key := fmt.Sprintf("uv:%s:%s", page, date)
    
    // 使用HyperLogLog记录唯一访客
    if err := s.redisClient.PFAdd(ctx, key, visitorID).Err(); err != nil {
        return fmt.Errorf("failed to record unique visitor: %w", err)
    }
    
    return nil
}
```

### 2. 统计收集器 (StatsCollector)

提供更高级别的统计功能，简化对StatsService的调用：

- **一站式记录**：同时记录PV和UV
- **日期聚合**：支持按日期范围查询统计数据
- **便捷查询**：提供当天统计快速查询功能

```go
// 同时记录一次页面访问的PV和UV
func (c *StatsCollector) RecordVisit(ctx context.Context, page, visitorID string) error {
    // 记录PV
    if err := c.service.RecordPageView(ctx, page); err != nil {
        return fmt.Errorf("failed to record page view: %w", err)
    }

    // 记录UV
    if err := c.service.RecordUniqueVisitor(ctx, page, visitorID); err != nil {
        return fmt.Errorf("failed to record unique visitor: %w", err)
    }

    return nil
}
```

### 3. HTTP处理器 (StatsHandler)

提供HTTP API接口，处理客户端请求：

- **记录访问**：接收并处理页面访问记录
- **查询功能**：提供多种统计数据查询接口
- **路由管理**：配置和管理所有HTTP路由

```go
// 处理记录页面访问的请求
func (h *StatsHandler) RecordVisit(c *gin.Context) {
    var req struct {
        Page      string `json:"page" binding:"required"`
        VisitorID string `json:"visitor_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request parameters: " + err.Error(),
        })
        return
    }

    if err := h.collector.RecordVisit(c.Request.Context(), req.Page, req.VisitorID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to record visit: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "Visit recorded successfully",
    })
}
```

## UV/PV统计原理

### 1. PV (页面浏览量) 统计

系统使用Redis的计数器功能统计PV：

- **存储结构**：每个页面每天使用一个独立的计数器
- **键命名**：使用格式 `pv:{page}:{date}` 确保唯一性
- **操作命令**：使用Redis的INCR命令原子递增计数
- **数据查询**：使用GET命令获取特定页面和日期的PV值

PV统计逻辑示意：
```
1. 构造键名 pv:/home:2025-04-21
2. 对该键执行INCR操作
3. 计数器值增加1，表示新增一次页面访问
```

### 2. UV (唯一访客) 统计

系统使用Redis的HyperLogLog数据结构高效统计UV：

- **存储结构**：每个页面每天使用一个HyperLogLog结构
- **键命名**：使用格式 `uv:{page}:{date}` 确保唯一性
- **操作命令**：使用PFADD添加访客标识，PFCOUNT获取基数估计
- **空间效率**：无论元素数量多少，只占用固定空间（12KB左右）

UV统计逻辑示意：
```
1. 构造键名 uv:/home:2025-04-21
2. 对该键执行PFADD操作，添加访客ID (如user_123)
3. 如果访客ID已存在，UV计数不变
4. 如果访客ID是新的，UV计数增加
```

### 3. 日期范围统计

系统支持获取日期范围内的统计数据：

1. **PV汇总**：循环获取每天的PV并累加
2. **UV汇总**：循环获取每天的UV并累加
    - 注：此方法在跨天访客重复时会重复计数，实际场景可能需要更复杂的合并逻辑

## 如何运行系统

### 前提条件

- Go 1.16或更高版本
- Redis 5.0或更高版本（需支持HyperLogLog）

### 配置系统

1. **Redis连接配置**：
   系统默认使用本地Redis实例，无需密码。如需修改，请更新`internal/config/config.go`中的配置：

   ```go
   // DefaultConfig 返回默认配置
   func DefaultConfig() *Config {
       return &Config{
           RedisAddr:     "localhost:6379",
           RedisPassword: "",
           RedisDB:       0,
           ServerAddr:    ":8080",
       }
   }
   ```

2. **确保Redis可访问**：
    - Redis服务需要正常运行
    - 确保应用有权限访问配置的Redis实例

### 启动步骤

1. **构建并运行程序**：
   ```bash
   cd uv-pv-collector
   go build -o uv-pv-collector cmd/main.go
   ./uv-pv-collector
   ```

   或者直接运行：
   ```bash
   cd uv-pv-collector
   go run cmd/main.go
   ```

2. **观察系统日志**：
   程序启动后会显示HTTP服务器启动信息。

### 测试系统功能

1. **记录页面访问**：
   ```bash
   curl -X POST http://localhost:8080/record \
     -H "Content-Type: application/json" \
     -d '{"page":"/home", "visitor_id":"user1"}'
   ```

2. **获取今日统计数据**：
   ```bash
   curl "http://localhost:8080/stats/today?page=/home"
   ```

3. **获取特定日期统计数据**：
   ```bash
   curl "http://localhost:8080/stats/daily?page=/home&date=2025-04-21"
   ```

4. **获取日期范围统计数据**：
   ```bash
   curl "http://localhost:8080/stats/range?page=/home&start_date=2025-04-20&end_date=2025-04-21"
   ```

5. **健康检查**：
   ```bash
   curl http://localhost:8080/ping
   ```

## 代码结构

- `cmd/`: 应用入口
    - main.go: 主程序，初始化并协调各组件

- `internal/`: 内部实现
    - `config/`: 配置管理
        - config.go: Redis连接和服务器配置
    - `stats/`: 统计功能实现
        - service.go: Redis操作封装，提供PV和UV底层功能
        - collector.go: 高级统计服务，提供便捷的统计方法
    - `handlers/`: HTTP处理
        - stats_handler.go: HTTP请求处理器，提供Web API

- `go.mod`: Go模块定义文件