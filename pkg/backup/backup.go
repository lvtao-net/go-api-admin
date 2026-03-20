package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/config"
)

// BackupService 备份服务
type BackupService struct {
	backupPath string
	retainDays int
	enabled    bool
}

// NewBackupService 创建备份服务
func NewBackupService() *BackupService {
	cfg := config.GetBackupSettings()

	backupPath, _ := cfg["path"].(string)
	if backupPath == "" {
		backupPath = "./backups"
	}
	retainDays, _ := cfg["retainDays"].(int)
	if retainDays == 0 {
		retainDays = 5
	}
	enabled, _ := cfg["enabled"].(bool)

	return &BackupService{
		backupPath: backupPath,
		retainDays: retainDays,
		enabled:    enabled,
	}
}

// BackupResult 备份结果
type BackupResult struct {
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
}

// BackupDatabase 备份数据库
func (s *BackupService) BackupDatabase() (*BackupResult, error) {
	cfg := config.GetConfig()
	dbCfg := cfg.Database

	// 创建备份目录
	if err := os.MkdirAll(s.backupPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(s.backupPath, filename)

	// 使用 mysqldump 备份
	cmd := exec.Command(
		"mysqldump",
		"-h", dbCfg.Host,
		"-P", fmt.Sprintf("%d", dbCfg.Port),
		"-u", dbCfg.User,
		"-p"+dbCfg.Password,
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
		dbCfg.Name,
	)

	output, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer output.Close()

	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		os.Remove(filepath)
		return nil, fmt.Errorf("failed to backup database: %w", err)
	}

	// 获取文件信息
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	return &BackupResult{
		Filename:  filename,
		Path:      filepath,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// RestoreDatabase 恢复数据库
func (s *BackupService) RestoreDatabase(backupFile string) error {
	cfg := config.GetConfig()
	dbCfg := cfg.Database

	// 检查备份文件是否存在
	if _, err := os.Stat(backupFile); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// 使用 mysql 恢复
	cmd := exec.Command(
		"mysql",
		"-h", dbCfg.Host,
		"-P", fmt.Sprintf("%d", dbCfg.Port),
		"-u", dbCfg.User,
		"-p"+dbCfg.Password,
		dbCfg.Name,
	)

	input, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer input.Close()

	cmd.Stdin = input
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	return nil
}

// ListBackups 列出所有备份
func (s *BackupService) ListBackups() ([]BackupResult, error) {
	entries, err := os.ReadDir(s.backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupResult
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) < 4 || entry.Name()[:4] != "back" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupResult{
			Filename:    entry.Name(),
			Path:        filepath.Join(s.backupPath, entry.Name()),
			Size:        info.Size(),
			CreatedAt:   info.ModTime(),
		})
	}

	return backups, nil
}

// DeleteBackup 删除备份
func (s *BackupService) DeleteBackup(backupFile string) error {
	return os.Remove(backupFile)
}

// CleanOldBackups 清理过期备份
func (s *BackupService) CleanOldBackups() error {
	backups, err := s.ListBackups()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -s.retainDays)

	for _, backup := range backups {
		if backup.CreatedAt.Before(cutoffTime) {
			if err := s.DeleteBackup(backup.Path); err != nil {
				continue
			}
		}
	}

	return nil
}

// StartAutoBackup 启动自动备份
func (s *BackupService) StartAutoBackup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			s.CleanOldBackups()
			s.BackupDatabase()
		}
	}()
}
