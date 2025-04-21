package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"uv-pv-collector/internal/stats"
)

// StatsHandler 处理与PV和UV统计相关的HTTP请求
type StatsHandler struct {
	collector *stats.StatsCollector
}

// NewStatsHandler 创建一个新的统计处理器
func NewStatsHandler(collector *stats.StatsCollector) *StatsHandler {
	return &StatsHandler{
		collector: collector,
	}
}

// Setup 设置所有路由
func (h *StatsHandler) Setup(router *gin.Engine) {
	// 记录访问
	router.POST("/record", h.RecordVisit)

	// 获取统计数据的路由
	statsApi := router.Group("/stats")
	{
		// 获取特定日期的统计数据
		statsApi.GET("/daily", h.GetDailyStats)
		// 获取今天的统计数据
		statsApi.GET("/today", h.GetTodayStats)
		// 获取日期范围内的统计数据
		statsApi.GET("/range", h.GetStatsForDateRange)
	}
}

// RecordVisit 处理记录页面访问的请求
func (h *StatsHandler) RecordVisit(c *gin.Context) {
	// 定义请求体结构
	var req struct {
		Page      string `json:"page" binding:"required"`
		VisitorID string `json:"visitor_id" binding:"required"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// 记录访问
	if err := h.collector.RecordVisit(c.Request.Context(), req.Page, req.VisitorID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record visit: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Visit recorded successfully",
	})
}

// GetDailyStats 处理获取特定日期统计数据的请求
func (h *StatsHandler) GetDailyStats(c *gin.Context) {
	page := c.Query("page")
	date := c.Query("date")

	if page == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Page and date parameters are required",
		})
		return
	}

	pv, uv, err := h.collector.GetDailyStats(c.Request.Context(), page, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":            page,
		"date":            date,
		"page_views":      pv,
		"unique_visitors": uv,
	})
}

// GetTodayStats 处理获取今天统计数据的请求
func (h *StatsHandler) GetTodayStats(c *gin.Context) {
	page := c.Query("page")

	if page == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Page parameter is required",
		})
		return
	}

	pv, uv, err := h.collector.GetTodayStats(c.Request.Context(), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get today's stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":            page,
		"page_views":      pv,
		"unique_visitors": uv,
	})
}

// GetStatsForDateRange 处理获取日期范围内统计数据的请求
func (h *StatsHandler) GetStatsForDateRange(c *gin.Context) {
	page := c.Query("page")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page == "" || startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Page, start_date and end_date parameters are required",
		})
		return
	}

	pv, uv, err := h.collector.GetStatsForDateRange(c.Request.Context(), page, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stats for date range: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":                  page,
		"start_date":            startDate,
		"end_date":              endDate,
		"total_page_views":      pv,
		"total_unique_visitors": uv,
		"note":                  "UV count across multiple days may count some visitors multiple times",
	})
}
