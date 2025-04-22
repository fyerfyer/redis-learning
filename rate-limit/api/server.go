package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rate-limit/pkg/cache"
	"rate-limit/pkg/detector"
	"rate-limit/pkg/limiter"
	"rate-limit/pkg/storage"
)

// Server API服务器
type Server struct {
	redisClient *storage.RedisClient
	localCache  *cache.LocalCache
	hotKeyDet   *detector.HotKeyDetector
	rateLimiter *limiter.RateLimiter
	router      *gin.Engine
	port        string
}

// NewServer 创建一个新的API服务器
func NewServer(port string) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		redisClient: storage.NewRedisClient(),
		localCache:  cache.NewLocalCache(5*time.Minute, time.Minute),
		hotKeyDet:   detector.NewDefaultHotKeyDetector(),
		rateLimiter: limiter.NewDefaultRateLimiter(),
		router:      gin.Default(),
		port:        port,
	}

	s.setupRoutes()
	return s
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	s.router.GET("/get/:key", s.handleGetKey)
	s.router.GET("/stats/:key", s.handleKeyStats)
	s.router.GET("/hot-keys", s.handleHotKeys)
	s.router.POST("/set/:key", s.handleSetKey)
}

// Start 启动服务器
func (s *Server) Start() error {
	log.Printf("Starting API server on port %s", s.port)
	return s.router.Run(":" + s.port)
}

// handleGetKey 处理获取key的请求
func (s *Server) handleGetKey(c *gin.Context) {
	key := c.Param("key")

	// 记录访问并检测是否为热点key
	isHotKey := s.hotKeyDet.RecordAccess(key)

	// 如果是热点key，应用限流
	if isHotKey {
		// 尝试从本地缓存获取
		if value, found := s.localCache.Get(key); found {
			log.Printf("Hot key cache hit: %s", key)
			c.JSON(http.StatusOK, gin.H{"value": value, "source": "local_cache"})
			return
		}

		// 如果本地缓存没有，检查是否允许访问Redis
		if !s.rateLimiter.Allow(key) {
			log.Printf("Rate limited for hot key: %s", key)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests for this hot key"})
			return
		}
	}

	// 从Redis获取数据
	value, err := s.redisClient.Get(key)
	if err != nil {
		log.Printf("Error getting key from Redis: %s, %v", key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get value from Redis"})
		return
	}

	// 如果为空，表示key不存在
	if value == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	// 如果是热点key，更新本地缓存
	if isHotKey {
		s.localCache.Set(key, value, 5*time.Minute)
		log.Printf("Hot key cached: %s", key)
	}

	c.JSON(http.StatusOK, gin.H{"value": value, "source": "redis"})
}

// handleKeyStats 获取key的统计信息
func (s *Server) handleKeyStats(c *gin.Context) {
	key := c.Param("key")

	accessCount := s.hotKeyDet.GetAccessCount(key)
	isHotKey := s.hotKeyDet.IsHotKey(key)
	inCache, _ := s.localCache.Get(key)

	c.JSON(http.StatusOK, gin.H{
		"key":          key,
		"access_count": accessCount,
		"is_hot_key":   isHotKey,
		"in_cache":     inCache != "",
	})
}

// handleHotKeys 获取当前热点key列表
func (s *Server) handleHotKeys(c *gin.Context) {
	// 获取热点key列表
	hotKeys := s.hotKeyDet.GetHotKeys()
	c.JSON(http.StatusOK, gin.H{"hot_keys": hotKeys})
}

// handleSetKey 设置key的值
func (s *Server) handleSetKey(c *gin.Context) {
	key := c.Param("key")
	value := c.PostForm("value")

	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value cannot be empty"})
		return
	}

	// 设置到Redis
	expiration := 1 * time.Hour // 默认过期时间1小时
	err := s.redisClient.Set(key, value, expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set value in Redis"})
		return
	}

	// 如果是热点key，也更新本地缓存
	if s.hotKeyDet.IsHotKey(key) {
		s.localCache.Set(key, value, 5*time.Minute)
		log.Printf("Hot key cache updated: %s", key)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Close 关闭服务器和相关资源
func (s *Server) Close() {
	if s.redisClient != nil {
		err := s.redisClient.Close()
		if err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}
}
