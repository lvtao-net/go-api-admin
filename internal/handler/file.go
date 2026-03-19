package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type FileHandler struct {
	uploadPath string
	maxSize    int64
}

func NewFileHandler() *FileHandler {
	settingService := service.NewSettingService()
	uploadPath := settingService.Get("upload.path")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	maxFileSize := settingService.GetInt("upload.maxFileSize", 10)

	return &FileHandler{
		uploadPath: uploadPath,
		maxSize:    int64(maxFileSize) * 1024 * 1024,
	}
}

// Upload 上传文件
func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}

	// 验证文件大小
	if file.Size > h.maxSize {
		response.BadRequest(c, fmt.Sprintf("File size exceeds limit of %d MB", h.maxSize/1024/1024))
		return
	}

	// 验证文件类型（使用默认允许的类型）
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "application/pdf"}
	contentType := file.Header.Get("Content-Type")
	allowed := false
	for _, t := range allowedTypes {
		if t == contentType {
			allowed = true
			break
		}
	}
	if !allowed {
		response.BadRequest(c, "File type not allowed")
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s%s", time.Now().UnixNano(), ext)

	// 创建目录
	datePath := time.Now().Format("2006/01/02")
	fullPath := filepath.Join(h.uploadPath, datePath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		response.InternalError(c, "Failed to create upload directory")
		return
	}

	// 保存文件
	dst := filepath.Join(fullPath, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	// 返回文件信息
	fileURL := fmt.Sprintf("/api/files/%s/%s", datePath, filename)
	response.Success(c, gin.H{
		"filename": filename,
		"url":      fileURL,
		"size":     file.Size,
		"type":     contentType,
	})
}

// UploadMultiple 上传多个文件
func (h *FileHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "No files uploaded")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.BadRequest(c, "No files uploaded")
		return
	}

	// 使用默认允许的类型
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "application/pdf"}
	var results []map[string]interface{}

	for _, file := range files {
		// 验证文件大小
		if file.Size > h.maxSize {
			continue
		}

		// 验证文件类型
		contentType := file.Header.Get("Content-Type")
		allowed := false
		for _, t := range allowedTypes {
			if t == contentType {
				allowed = true
				break
			}
		}
		if !allowed {
			continue
		}

		// 生成唯一文件名
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d%s%s", time.Now().UnixNano(), ext)

		// 创建目录
		datePath := time.Now().Format("2006/01/02")
		fullPath := filepath.Join(h.uploadPath, datePath)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			continue
		}

		// 保存文件
		dst := filepath.Join(fullPath, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			continue
		}

		// 添加到结果
		fileURL := fmt.Sprintf("/api/files/%s/%s", datePath, filename)
		results = append(results, map[string]interface{}{
			"filename": filename,
			"url":      fileURL,
			"size":     file.Size,
			"type":     contentType,
		})
	}

	response.Success(c, gin.H{"files": results})
}

// GetFile 获取文件
func (h *FileHandler) GetFile(c *gin.Context) {
	filename := c.Param("filename")

	// 安全检查：防止路径遍历
	if filename == "" || filename != filepath.Base(filename) {
		response.BadRequest(c, "Invalid filename")
		return
	}

	// Token 认证访问模式
	token := c.Query("token")
	if token != "" {
		// 验证 token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid token")
			return
		}
		_ = claims // 可以记录访问日志
	}

	// 构建文件路径
	filePath := filepath.Join(h.uploadPath, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.NotFound(c, "File not found")
		return
	}

	// 设置正确的 Content-Type
	contentType := getContentType(filename)
	c.Header("Content-Type", contentType)

	// 如果是图片等，设置缓存
	if isImage(filename) {
		c.Header("Cache-Control", "public, max-age=31536000")
	}

	c.File(filePath)
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile(c *gin.Context) {
	filename := c.Param("filename")
	
	// 安全检查：防止路径遍历
	if filename == "" || filename != filepath.Base(filename) {
		response.BadRequest(c, "Invalid filename")
		return
	}

	// 构建文件路径
	filePath := filepath.Join(h.uploadPath, filename)
	
	// 删除文件
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			response.NotFound(c, "File not found")
			return
		}
		response.InternalError(c, "Failed to delete file")
		return
	}

	response.Success(c, nil)
}

// getContentType 获取文件 Content-Type
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// isImage 判断是否为图片
func isImage(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return true
	default:
		return false
	}
}
