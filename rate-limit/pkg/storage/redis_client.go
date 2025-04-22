package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置参数
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DefaultConfig 默认Redis配置
var DefaultConfig = RedisConfig{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
}

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建一个新的Redis客户端
func NewRedisClient() *RedisClient {
	return NewRedisClientWithConfig(DefaultConfig)
}

// NewRedisClientWithConfig 使用指定配置创建Redis客户端
func NewRedisClientWithConfig(config RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 创建上下文
	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Printf("Successfully connected to Redis at %s", config.Addr)
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}
}

// Get 获取键值
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Printf("Error getting key %s: %v", key, err)
		return "", err
	}
	if errors.Is(err, redis.Nil) {
		return "", nil // 键不存在返回空字符串
	}
	return val, nil
}

// Set 设置键值
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Error setting key %s: %v", key, err)
	}
	return err
}

// Incr 递增键的值
func (r *RedisClient) Incr(key string) (int64, error) {
	val, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		log.Printf("Error incrementing key %s: %v", key, err)
	}
	return val, err
}

// SetNX 当key不存在时设置键值
func (r *RedisClient) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, expiration).Result()
}

// Del 删除键
func (r *RedisClient) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}
