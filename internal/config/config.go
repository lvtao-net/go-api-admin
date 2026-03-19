package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Admin    AdminConfig    `mapstructure:"admin"`
	Embed    EmbedConfig    `mapstructure:"embed"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Mode    string `mapstructure:"mode"`
	Port    int    `mapstructure:"port"`
}

type EmbedConfig struct {
	Enabled bool `mapstructure:"enabled"` // 是否使用嵌入的静态文件
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	LogMode      bool   `mapstructure:"log_mode"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Charset)
}

type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	Expires        int    `mapstructure:"expires"`
	RefreshExpires int    `mapstructure:"refresh_expires"`
}

type AdminConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

var GlobalConfig *Config

func Init(configPath string) error {
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 允许环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func GetConfig() *Config {
	return GlobalConfig
}

// GetAppSettings 获取应用设置（从数据库读取）
func GetAppSettings() map[string]interface{} {
	return map[string]interface{}{
		"logLevel":     "info",
		"logFormat":    "json",
		"uploadPath":   "./uploads",
		"maxFileSize":  10,
		"storageType": "local",
	}
}

// GetMailSettings 获取邮件设置
func GetMailSettings() map[string]interface{} {
	return map[string]interface{}{
		"enabled":   false,
		"host":      "",
		"port":      587,
		"username":  "",
		"password":  "",
		"fromEmail": "",
		"fromName":  "Go Gin API Admin",
	}
}

// GetRateLimitSettings 获取速率限制设置
func GetRateLimitSettings() map[string]interface{} {
	return map[string]interface{}{
		"enabled":           true,
		"requestsPerMinute": 60,
		"burst":            10,
	}
}

// GetBackupSettings 获取备份设置
func GetBackupSettings() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    false,
		"path":       "./backups",
		"retainDays": 5,
	}
}
