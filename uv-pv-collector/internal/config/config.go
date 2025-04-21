package config

// Config 存储Redis连接的配置信息
type Config struct {
	// Redis连接地址
	RedisAddr string
	// Redis密码，没有则为空
	RedisPassword string
	// Redis数据库索引
	RedisDB int
	// 应用服务器监听地址
	ServerAddr string
}

// DefaultConfig 返回默认配置
// 默认使用本地Redis，无密码，0号数据库
func DefaultConfig() *Config {
	return &Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		ServerAddr:    ":8080",
	}
}
