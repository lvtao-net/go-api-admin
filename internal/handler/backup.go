package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/pkg/backup"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type BackupHandler struct {
	backupSvc *backup.BackupService
}

func NewBackupHandler() *BackupHandler {
	return &BackupHandler{
		backupSvc: backup.NewBackupService(),
	}
}

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	Filename string `json:"filename"`
}

// CreateBackup 创建备份
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	result, err := h.backupSvc.BackupDatabase()
	if err != nil {
		response.InternalError(c, "Failed to create backup: "+err.Error())
		return
	}

	response.Success(c, result)
}

// RestoreBackup 恢复备份
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	var req struct {
		Filename string `json:"filename" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Filename is required")
		return
	}

	err := h.backupSvc.RestoreDatabase(req.Filename)
	if err != nil {
		response.InternalError(c, "Failed to restore backup: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Backup restored successfully"})
}

// ListBackups 列出备份
func (h *BackupHandler) ListBackups(c *gin.Context) {
	backups, err := h.backupSvc.ListBackups()
	if err != nil {
		response.InternalError(c, "Failed to list backups: "+err.Error())
		return
	}

	response.Success(c, gin.H{"items": backups, "total": len(backups)})
}

// DeleteBackup 删除备份
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.BadRequest(c, "Filename is required")
		return
	}

	err := h.backupSvc.DeleteBackup(filename)
	if err != nil {
		response.InternalError(c, "Failed to delete backup: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Backup deleted successfully"})
}
