package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"gorm.io/gorm"
)

// LogService 日志服务
type LogService struct {
	db         *gorm.DB
	logChan    chan model.OperationLog
	workerDone chan bool
}

var logService *LogService
var logOnce sync.Once

func NewLogService() *LogService {
	logOnce.Do(func() {
		logService = &LogService{
			db:         database.GetDB(),
			logChan:    make(chan model.OperationLog, 1000),
			workerDone: make(chan bool),
		}
		// 启动异步写入worker
		go logService.worker()
	})
	return logService
}

// worker 异步写入日志
func (s *LogService) worker() {
	for {
		select {
		case log := <-s.logChan:
			s.db.Create(&log)
		case <-s.workerDone:
			return
		}
	}
}

// Log 记录操作日志（异步）
func (s *LogService) Log(log model.OperationLog) {
	// 设置默认值（ID 自动递增，不需要手动设置）
	if log.Created.IsZero() {
		log.Created = time.Now()
	}
	log.Updated = log.Created

	// 发送到异步队列
	select {
	case s.logChan <- log:
	default:
		// 队列满了，丢弃日志
	}
}

// LogSync 同步记录日志（用于重要操作）
func (s *LogService) LogSync(log model.OperationLog) error {
	// ID 自动递增，不需要手动设置
	if log.Created.IsZero() {
		log.Created = time.Now()
	}
	log.Updated = log.Created
	return s.db.Create(&log).Error
}

// List 获取日志列表
func (s *LogService) List(page, perPage int, filters map[string]interface{}) ([]model.OperationLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}

	query := s.db.Model(&model.OperationLog{})

	// 应用过滤条件
	if collection, ok := filters["collection"].(string); ok && collection != "" {
		query = query.Where("collection = ?", collection)
	}
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if userType, ok := filters["userType"].(string); ok && userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if userID, ok := filters["userId"]; ok {
		switch v := userID.(type) {
		case uint64:
			query = query.Where("user_id = ?", v)
		case float64:
			query = query.Where("user_id = ?", uint64(v))
		case string:
			if v != "" {
				query = query.Where("user_id = ?", v)
			}
		}
	}
	if startDate, ok := filters["startDate"].(string); ok && startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created >= ?", t)
		}
	}
	if endDate, ok := filters["endDate"].(string); ok && endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created <= ?", t.Add(24*time.Hour))
		}
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * perPage
	var logs []model.OperationLog
	err := query.Order("created DESC").Offset(offset).Limit(perPage).Find(&logs).Error

	return logs, total, err
}

// GetByID 通过ID获取日志
func (s *LogService) GetByID(id uint64) (*model.OperationLog, error) {
	var log model.OperationLog
	if err := s.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// DeleteOldLogs 删除旧日志
func (s *LogService) DeleteOldLogs(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created < ?", cutoff).Delete(&model.OperationLog{}).Error
}

// GetStats 获取日志统计
func (s *LogService) GetStats(days int) (map[string]interface{}, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	// 按操作类型统计
	var byAction []struct {
		Action string
		Count  int64
	}
	s.db.Model(&model.OperationLog{}).
		Select("action, COUNT(*) as count").
		Where("created >= ?", cutoff).
		Group("action").
		Scan(&byAction)

	// 按用户类型统计
	var byUserType []struct {
		UserType string
		Count    int64
	}
	s.db.Model(&model.OperationLog{}).
		Select("user_type, COUNT(*) as count").
		Where("created >= ?", cutoff).
		Group("user_type").
		Scan(&byUserType)

	// 按集合统计
	var byCollection []struct {
		Collection string
		Count      int64
	}
	s.db.Model(&model.OperationLog{}).
		Select("collection, COUNT(*) as count").
		Where("created >= ?", cutoff).
		Group("collection").
		Order("count DESC").
		Limit(10).
		Scan(&byCollection)

	// 今日统计
	today := time.Now().Format("2006-01-02")
	var todayCount int64
	s.db.Model(&model.OperationLog{}).
		Where("created >= ?", today).
		Count(&todayCount)

	stats := map[string]interface{}{
		"byAction":      byAction,
		"byUserType":    byUserType,
		"byCollection":  byCollection,
		"todayCount":    todayCount,
		"totalCount":    days,
	}

	return stats, nil
}

// Close 关闭日志服务
func (s *LogService) Close() {
	close(s.workerDone)
}

// OperationLogRequest 创建日志请求
type OperationLogRequest struct {
	Collection string      `json:"collection,omitempty"`
	RecordID   uint64      `json:"recordId,omitempty"`
	Action     string      `json:"action"`
	UserID     uint64      `json:"userId,omitempty"`
	UserEmail  string      `json:"userEmail,omitempty"`
	UserType   string      `json:"userType,omitempty"`
	IP         string      `json:"ip,omitempty"`
	UserAgent  string      `json:"userAgent,omitempty"`
	Request    interface{} `json:"request,omitempty"`
	Response   interface{} `json:"response,omitempty"`
	Status     int         `json:"status"`
}

// ToModel 转换为日志模型
func (r *OperationLogRequest) ToModel() model.OperationLog {
	reqBytes, _ := json.Marshal(r.Request)
	resBytes, _ := json.Marshal(r.Response)

	return model.OperationLog{
		Collection: r.Collection,
		RecordID:   r.RecordID,
		Action:     r.Action,
		UserID:     r.UserID,
		UserEmail:  r.UserEmail,
		UserType:   r.UserType,
		IP:         r.IP,
		UserAgent:  r.UserAgent,
		Request:    string(reqBytes),
		Response:   string(resBytes),
		Status:     r.Status,
	}
}
