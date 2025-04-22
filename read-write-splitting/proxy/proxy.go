package proxy

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"read-write-splitting/internal/config"
)

var (
	// ErrNoSlaveAvailable 当没有可用的从库时返回此错误
	ErrNoSlaveAvailable = errors.New("no slave available")
)

// RedisProxy Redis读写分离代理
type RedisProxy struct {
	master      *redis.Client              // 主库连接
	slaves      []*redis.Client            // 从库连接列表
	balancer    Balancer                   // 负载均衡器
	commandType map[string]bool            // 命令类型映射表，true表示写命令，false表示读命令
	config      *config.RedisClusterConfig // Redis集群配置
}

// NewRedisProxy 创建一个新的Redis读写分离代理
func NewRedisProxy(cfg *config.RedisClusterConfig) *RedisProxy {
	// 初始化主库连接
	master := redis.NewClient(&redis.Options{
		Addr:     cfg.GetMasterAddress(),
		Password: cfg.Master.Password,
		DB:       cfg.Master.DB,
		PoolSize: cfg.PoolSize,
	})

	// 初始化从库连接列表
	slaves := make([]*redis.Client, len(cfg.Slaves))
	for i, slaveCfg := range cfg.Slaves {
		slaves[i] = redis.NewClient(&redis.Options{
			Addr:     slaveCfg.Host + ":" + strconv.Itoa(slaveCfg.Port),
			Password: slaveCfg.Password,
			DB:       slaveCfg.DB,
			PoolSize: cfg.PoolSize / len(cfg.Slaves), // 将连接池均匀分配给从库
		})
	}

	// 初始化负载均衡器
	balancer := NewRoundRobinBalancer(len(slaves))

	return &RedisProxy{
		master:   master,
		slaves:   slaves,
		balancer: balancer,
		config:   cfg,
		commandType: map[string]bool{
			// 写命令
			"set":    true,
			"setex":  true,
			"setnx":  true,
			"del":    true,
			"incr":   true,
			"decr":   true,
			"expire": true,
			"lpush":  true,
			"rpush":  true,
			"sadd":   true,
			"zadd":   true,
			"hset":   true,
			// 读命令
			"get":       false,
			"mget":      false,
			"exists":    false,
			"lrange":    false,
			"lindex":    false,
			"smembers":  false,
			"sismember": false,
			"zrange":    false,
			"hget":      false,
			"hgetall":   false,
		},
	}
}

// Close 关闭所有Redis连接
func (p *RedisProxy) Close() error {
	var err error

	// 关闭主库连接
	if cerr := p.master.Close(); cerr != nil {
		err = cerr
	}

	// 关闭所有从库连接
	for _, slave := range p.slaves {
		if cerr := slave.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}

	return err
}

// IsWriteCommand 判断命令是否为写命令
func (p *RedisProxy) IsWriteCommand(cmd string) bool {
	isWrite, exists := p.commandType[cmd]
	if !exists {
		// 默认未知命令路由到主库
		return true
	}
	return isWrite
}

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

// processOnMaster 在主库上处理命令
func (p *RedisProxy) processOnMaster(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	// 使用主库执行命令
	return p.master.Do(ctx, cmd, args).Result()
}

// processOnSlave 在从库上处理命令
func (p *RedisProxy) processOnSlave(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	// 从负载均衡器获取从库索引
	slaveIndex := p.balancer.Next(len(p.slaves))
	if slaveIndex < 0 {
		// 没有可用从库，降级到主库
		fmt.Println("No slave available, falling back to master")
		return p.processOnMaster(ctx, cmd, args...)
	}

	// 选择从库执行命令
	result, err := p.slaves[slaveIndex].Do(ctx, cmd, args).Result()
	if err != nil {
		// 从库出错，标记为不可用
		p.balancer.MarkDown(slaveIndex)

		// 尝试重新选择从库
		slaveIndex = p.balancer.Next(len(p.slaves))
		if slaveIndex < 0 {
			// 没有更多可用从库，降级到主库
			fmt.Println("Slave failed, falling back to master")
			return p.processOnMaster(ctx, cmd, args...)
		}

		// 在另一个从库上重试
		result, err = p.slaves[slaveIndex].Do(ctx, cmd, args).Result()
		if err != nil {
			// 第二次尝试也失败，降级到主库
			p.balancer.MarkDown(slaveIndex)
			fmt.Println("Second slave failed, falling back to master")
			return p.processOnMaster(ctx, cmd, args...)
		}
	}

	return result, err
}

// HealthCheck 执行健康检查，恢复标记为不可用的从库
func (p *RedisProxy) HealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i, slave := range p.slaves {
		// 尝试ping从库
		_, err := slave.Ping(ctx).Result()
		if err == nil {
			// 从库可用，标记为可用
			p.balancer.MarkUp(i)
		} else {
			// 从库不可用，标记为不可用
			p.balancer.MarkDown(i)
			fmt.Printf("Slave %d is down: %v\n", i, err)
		}
	}
}

// StartHealthCheck 开始定期健康检查
func (p *RedisProxy) StartHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			p.HealthCheck()
		}
	}()
}
