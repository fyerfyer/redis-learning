package config

// RedisConfig 定义单个Redis实例的配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// RedisClusterConfig 定义Redis读写分离集群的配置
type RedisClusterConfig struct {
	Master   RedisConfig   // 主库配置
	Slaves   []RedisConfig // 从库配置列表
	PoolSize int           // 连接池大小
}

// DefaultConfig 返回默认的Redis集群配置
func DefaultConfig() *RedisClusterConfig {
	return &RedisClusterConfig{
		Master: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Slaves: []RedisConfig{
			{
				Host:     "localhost",
				Port:     6380,
				Password: "",
				DB:       0,
			},
			{
				Host:     "localhost",
				Port:     6381,
				Password: "",
				DB:       0,
			},
		},
		PoolSize: 10,
	}
}

// GetMasterAddress 获取主库地址
func (c *RedisClusterConfig) GetMasterAddress() string {
	return c.Master.GetAddress()
}

// GetSlaveAddresses 获取所有从库地址
func (c *RedisClusterConfig) GetSlaveAddresses() []string {
	addresses := make([]string, len(c.Slaves))
	for i, slave := range c.Slaves {
		addresses[i] = slave.GetAddress()
	}
	return addresses
}

// GetAddress 获取Redis实例的地址
func (c *RedisConfig) GetAddress() string {
	return c.Host + ":" + string(rune(c.Port+'0'))
}
