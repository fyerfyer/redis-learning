package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"uv-pv-collector/internal/config"
	"uv-pv-collector/internal/handlers"
	"uv-pv-collector/internal/stats"
)

func main() {
	// 加载配置
	cfg := config.DefaultConfig()

	// 初始化StatsService
	statsService, err := stats.NewStatsService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize stats service: %v", err)
	}
	defer statsService.Close()

	// 初始化StatsCollector
	collector := stats.NewStatsCollector(statsService)

	// 初始化Gin路由器
	router := gin.Default()

	// 添加一个简单的健康检查路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 设置统计处理器路由
	statsHandler := handlers.NewStatsHandler(collector)
	statsHandler.Setup(router)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("Server starting on %s", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
