package config

import (
	"time"
)

// Config 系统总体配置
type Config struct {
	// Redis相关配置
	Redis RedisConfig

	// 本地缓存相关配置
	LocalCache LocalCacheConfig

	// 多级缓存配置
	MultiLevelCache MultiLevelCacheConfig
}

// RedisConfig Redis配置
type RedisConfig struct {
	// Redis服务器地址
	Addr string

	// Redis密码，可为空
	Password string

	// Redis数据库索引
	DB int

	// 连接池大小
	PoolSize int

	// 连接超时时间
	DialTimeout time.Duration

	// 读超时
	ReadTimeout time.Duration

	// 写超时
	WriteTimeout time.Duration
}

// LocalCacheConfig 本地缓存配置
type LocalCacheConfig struct {
	// 缓存最大条目数
	MaxEntries int

	// 默认过期时间
	DefaultExpiration time.Duration

	// 清除过期数据的检查周期
	CleanupInterval time.Duration
}

// MultiLevelCacheConfig 多级缓存配置
type MultiLevelCacheConfig struct {
	// 本地缓存的过期时间系数（相对于Redis中的过期时间）
	// 例如：0.5表示本地缓存过期时间为Redis过期时间的一半
	LocalExpirationFactor float64

	// 是否启用热点key检测
	EnableHotKeyDetection bool

	// 访问频率阈值，超过此值视为热点key
	HotKeyThreshold int64

	// 热点key统计时间窗口
	HotKeyWindow time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Redis: RedisConfig{
			Addr:         "localhost:6379",
			Password:     "",
			DB:           0,
			PoolSize:     10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		LocalCache: LocalCacheConfig{
			MaxEntries:        1000,
			DefaultExpiration: 5 * time.Minute,
			CleanupInterval:   10 * time.Minute,
		},
		MultiLevelCache: MultiLevelCacheConfig{
			LocalExpirationFactor: 0.5,
			EnableHotKeyDetection: true,
			HotKeyThreshold:       100,
			HotKeyWindow:          1 * time.Minute,
		},
	}
}
