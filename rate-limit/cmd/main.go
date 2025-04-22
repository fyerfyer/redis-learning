package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rate-limit/api"
)

func main() {
	log.Printf("Starting hot key detection and rate limiting system...")

	// 创建并启动API服务器
	server := api.NewServer("8080")

	// 优雅关闭处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器（非阻塞）
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Failed to start server: %v", err)
			quit <- syscall.SIGTERM
		}
	}()

	log.Println("Rate limiting server is running on port 8080")
	log.Printf("Press Ctrl+C to shut down")

	// 等待关闭信号
	<-quit
	log.Printf("Shutting down server...")

	// 关闭资源
	server.Close()
	log.Printf("Server stopped")

	// 给日志一点时间写入
	time.Sleep(time.Millisecond * 100)
}
