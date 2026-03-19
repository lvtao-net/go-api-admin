package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

// LogHandler 日志处理器
type LogHandler struct {
	logService *service.LogService
}

func NewLogHandler() *LogHandler {
	return &LogHandler{
		logService: service.NewLogService(),
	}
}

// List 获取操作日志列表
func (h *LogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))

	// 构建过滤条件
	filters := make(map[string]interface{})
	if collection := c.Query("collection"); collection != "" {
		filters["collection"] = collection
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if userType := c.Query("userType"); userType != "" {
		filters["userType"] = userType
	}
	if userId := c.Query("userId"); userId != "" {
		filters["userId"] = userId
	}
	if startDate := c.Query("startDate"); startDate != "" {
		filters["startDate"] = startDate
	}
	if endDate := c.Query("endDate"); endDate != "" {
		filters["endDate"] = endDate
	}

	logs, total, err := h.logService.List(page, perPage, filters)
	if err != nil {
		response.Error(c, 500, "获取日志失败")
		return
	}

	totalPages := (int(total) + perPage - 1) / perPage

	response.Success(c, gin.H{
		"page":        page,
		"perPage":     perPage,
		"totalItems":  total,
		"totalPages":  totalPages,
		"items":       logs,
	})
}

// GetStats 获取日志统计
func (h *LogHandler) GetStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	stats, err := h.logService.GetStats(days)
	if err != nil {
		response.Error(c, 500, "获取统计失败")
		return
	}

	response.Success(c, stats)
}

// DeleteOldLogs 清理旧日志
func (h *LogHandler) DeleteOldLogs(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	err := h.logService.DeleteOldLogs(days)
	if err != nil {
		response.Error(c, 500, "清理日志失败")
		return
	}

	response.Success(c, gin.H{"deleted": days})
}
