package stats

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"uv-pv-collector/internal/config"
)

// StatsService 提供UV和PV统计的服务
type StatsService struct {
	redisClient *redis.Client
}

// NewStatsService 创建一个新的统计服务实例
func NewStatsService(cfg *config.Config) (*StatsService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试Redis连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &StatsService{
		redisClient: client,
	}, nil
}

// RecordPageView 记录页面浏览量(PV)
func (s *StatsService) RecordPageView(ctx context.Context, page string) error {
	date := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("pv:%s:%s", page, date)

	// 使用INCR命令增加计数器
	if err := s.redisClient.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to record page view: %w", err)
	}

	return nil
}

// RecordUniqueVisitor 记录唯一访客(UV)
func (s *StatsService) RecordUniqueVisitor(ctx context.Context, page, visitorID string) error {
	date := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("uv:%s:%s", page, date)

	// 使用HyperLogLog记录唯一访客
	if err := s.redisClient.PFAdd(ctx, key, visitorID).Err(); err != nil {
		return fmt.Errorf("failed to record unique visitor: %w", err)
	}

	return nil
}

// GetPageViews 获取特定页面在指定日期的PV数
func (s *StatsService) GetPageViews(ctx context.Context, page, date string) (int64, error) {
	key := fmt.Sprintf("pv:%s:%s", page, date)

	val, err := s.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		// 键不存在，返回0
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get page views: %w", err)
	}

	return val, nil
}

// GetUniqueVisitors 获取特定页面在指定日期的UV数
func (s *StatsService) GetUniqueVisitors(ctx context.Context, page, date string) (int64, error) {
	key := fmt.Sprintf("uv:%s:%s", page, date)

	val, err := s.redisClient.PFCount(ctx, key).Result()
	if err == redis.Nil {
		// 键不存在，返回0
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get unique visitors: %w", err)
	}

	return val, nil
}

// Close 关闭Redis连接
func (s *StatsService) Close() error {
	return s.redisClient.Close()
}
