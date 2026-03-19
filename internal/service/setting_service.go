package service

import (
	"encoding/json"
	"sync"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"gorm.io/gorm"
)

// SettingService 系统设置服务
type SettingService struct {
	db *gorm.DB
	// 缓存设置
	cache map[string]string
	mu    sync.RWMutex
}

var settingService *SettingService
var settingOnce sync.Once

func NewSettingService() *SettingService {
	settingOnce.Do(func() {
		settingService = &SettingService{
			db:    database.GetDB(),
			cache: make(map[string]string),
		}
		settingService.loadCache()
	})
	return settingService
}

// loadCache 从数据库加载设置到缓存
func (s *SettingService) loadCache() {
	var settings []model.Setting
	s.db.Find(&settings)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, setting := range settings {
		s.cache[setting.Key] = setting.Value
	}
}

// Get 获取设置值
func (s *SettingService) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

// GetInt 获取整数设置
func (s *SettingService) GetInt(key string, defaultVal int) int {
	val := s.Get(key)
	if val == "" {
		return defaultVal
	}
	var result int
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return defaultVal
	}
	return result
}

// GetBool 获取布尔设置
func (s *SettingService) GetBool(key string, defaultVal bool) bool {
	val := s.Get(key)
	if val == "" {
		return defaultVal
	}
	var result bool
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return defaultVal
	}
	return result
}

// GetMap 获取地图设置
func (s *SettingService) GetMap(key string) map[string]interface{} {
	val := s.Get(key)
	if val == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil
	}
	return result
}

// Set 设置值
func (s *SettingService) Set(key string, value interface{}) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case int, int64, float64:
		b, _ := json.Marshal(v)
		val = string(b)
	case bool:
		val = "true"
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		val = string(b)
	}

	// 更新或创建
	setting := &model.Setting{Key: key, Value: val}
	err := s.db.Where("`key` = ?", key).FirstOrCreate(setting, setting).Error
	if err != nil {
		return err
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[key] = val
	s.mu.Unlock()

	return nil
}

// Delete 删除设置
func (s *SettingService) Delete(key string) error {
	err := s.db.Where("`key` = ?", key).Delete(&model.Setting{}).Error
	if err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.cache, key)
	s.mu.Unlock()
	return nil
}

// GetAll 获取所有设置
func (s *SettingService) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range s.cache {
		var val interface{}
		if err := json.Unmarshal([]byte(v), &val); err != nil {
			val = v
		}
		result[k] = val
	}
	return result
}

// InitDefaultSettings 初始化默认设置
func (s *SettingService) InitDefaultSettings() error {
	defaults := map[string]interface{}{
		"app.name":                   "Go Gin API Admin",
		"app.logLevel":              "info",
		"app.logFormat":             "json",
		"upload.path":               "./uploads",
		"upload.maxFileSize":        10,
		"storage.type":              "local",
		"mail.enabled":             false,
		"mail.host":                "",
		"mail.port":                587,
		"mail.username":             "",
		"mail.password":            "",
		"mail.fromEmail":           "",
		"mail.fromName":            "Go Gin API Admin",
		"rateLimit.enabled":        true,
		"rateLimit.requestsPerMinute": 60,
		"rateLimit.burst":          10,
		"backup.enabled":          false,
		"backup.path":             "./backups",
		"backup.retainDays":        5,
	}

	for key, value := range defaults {
		if s.Get(key) == "" {
			if err := s.Set(key, value); err != nil {
				return err
			}
		}
	}

	return nil
}
