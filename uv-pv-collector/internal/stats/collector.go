package stats

import (
	"context"
	"fmt"
	"time"
)

// StatsCollector 统计数据收集器
// 提供了记录和查询网页访问数据的便捷方法
type StatsCollector struct {
	service *StatsService
}

// NewStatsCollector 创建一个新的统计收集器实例
func NewStatsCollector(service *StatsService) *StatsCollector {
	return &StatsCollector{
		service: service,
	}
}

// RecordVisit 同时记录一次页面访问的PV和UV
// page: 页面路径
// visitorID: 访客唯一标识(可以是IP, 用户ID等)
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

// GetDailyStats 获取指定页面某一天的PV和UV统计数据
func (c *StatsCollector) GetDailyStats(ctx context.Context, page, date string) (pv, uv int64, err error) {
	pv, err = c.service.GetPageViews(ctx, page, date)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get page views: %w", err)
	}

	uv, err = c.service.GetUniqueVisitors(ctx, page, date)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get unique visitors: %w", err)
	}

	return pv, uv, nil
}

// GetTodayStats 获取指定页面今天的PV和UV统计数据
func (c *StatsCollector) GetTodayStats(ctx context.Context, page string) (pv, uv int64, err error) {
	today := time.Now().Format("2006-01-02")
	return c.GetDailyStats(ctx, page, today)
}

// GetStatsForDateRange 获取指定页面在日期范围内的累计PV和UV
// startDate和endDate格式为"2006-01-02"
func (c *StatsCollector) GetStatsForDateRange(ctx context.Context, page, startDate, endDate string) (totalPV, totalUV int64, err error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end date format: %w", err)
	}

	// 收集日期范围内的PV总和
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		pv, err := c.service.GetPageViews(ctx, page, date)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get page views for %s: %w", date, err)
		}
		totalPV += pv
	}

	// 注意：这种方式统计UV不够准确，因为不同日期的UV可能有重复
	// 在实际生产环境中，可能需要使用更复杂的方法合并多个HyperLogLog
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		uv, err := c.service.GetUniqueVisitors(ctx, page, date)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get unique visitors for %s: %w", date, err)
		}
		totalUV += uv
	}

	return totalPV, totalUV, nil
}
