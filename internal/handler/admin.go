package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type AdminHandler struct {
	service *service.AdminService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		service: service.NewAdminService(),
	}
}

// parseUintID 将字符串 ID 转换为 uint64
func parseUintID(idStr string) (uint64, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format")
	}
	return id, nil
}

// GetStats 获取统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	db := database.GetDB()

	// 集合数量
	var collectionCount int64
	db.Table("_collections").Count(&collectionCount)

	// 记录数量（遍历所有集合）
	var totalRecords int64
	var collections []struct {
		Name string
		Type string
	}
	db.Table("_collections").Select("name, type").Where("type != ?", "view").Find(&collections)
	for _, col := range collections {
		var count int64
		db.Table(col.Name).Count(&count)
		totalRecords += count
	}

	// 管理员数量
	var adminCount int64
	db.Table("_admins").Count(&adminCount)

	// 计算存储空间
	var storageUsed int64
	db.Table("_files").Select("COALESCE(SUM(size), 0)").Scan(&storageUsed)

	// 格式化存储空间
	storageStr := formatStorageSize(storageUsed)

	response.Success(c, gin.H{
		"collections": collectionCount,
		"records":     totalRecords,
		"users":       adminCount,
		"storage":     storageStr,
	})
}

// formatStorageSize 格式化存储大小
func formatStorageSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	}
	if bytes >= MB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	}
	if bytes >= KB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%d B", bytes)
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	result, err := h.service.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, result)
}

// RefreshToken 刷新Token
func (h *AdminHandler) RefreshToken(c *gin.Context) {
	// 获取当前用户信息
	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	email, _ := c.Get("admin_email")

	// 生成新Token（adminID 可能是 string 或 uint，统一转为 string）
	var adminIDStr string
	switch v := adminID.(type) {
	case string:
		adminIDStr = v
	case uint:
		adminIDStr = fmt.Sprintf("%d", v)
	default:
		adminIDStr = fmt.Sprintf("%v", v)
	}

	token, err := auth.GenerateAdminToken(adminIDStr, email.(string))
	if err != nil {
		response.InternalError(c, "Failed to generate token")
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}

// GetProfile 获取管理员信息
func (h *AdminHandler) GetProfile(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// adminID 可能是 string 或 uint64
	var id uint64
	switch v := adminID.(type) {
	case string:
		var err error
		id, err = parseUintID(v)
		if err != nil {
			response.BadRequest(c, "Invalid admin ID")
			return
		}
	case uint64:
		id = v
	case float64:
		id = uint64(v)
	default:
		response.BadRequest(c, "Invalid admin ID type")
		return
	}

	admin, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "Admin not found")
		return
	}

	response.Success(c, gin.H{
		"id":      admin.ID,
		"email":   admin.Email,
		"created": admin.Created,
	})
}

// List 获取管理员列表
func (h *AdminHandler) List(c *gin.Context) {
	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}
	if pp := c.Query("perPage"); pp != "" {
		if _, err := fmt.Sscanf(pp, "%d", &perPage); err != nil {
			perPage = 20
		}
	}

	admins, total, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to get admins: "+err.Error())
		return
	}

	items := make([]gin.H, 0, len(admins))
	for _, admin := range admins {
		items = append(items, gin.H{
			"id":      admin.ID,
			"email":   admin.Email,
			"created": admin.Created,
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	response.Success(c, gin.H{
		"items":      items,
		"totalItems": total,
		"totalPages": totalPages,
		"page":       page,
		"perPage":    perPage,
	})
}

// Create 创建管理员
func (h *AdminHandler) Create(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	admin, err := h.service.CreateAdmin(req.Email, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":      admin.ID,
		"email":   admin.Email,
		"created": admin.Created,
	})
}

// Delete 删除管理员
func (h *AdminHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "Admin ID is required")
		return
	}

	id, err := parseUintID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid admin ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 更新管理员
func (h *AdminHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.BadRequest(c, "Admin ID is required")
		return
	}

	id, err := parseUintID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid admin ID")
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.service.Update(id, req.Email); err != nil {
		response.InternalError(c, "Failed to update admin: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改当前管理员密码
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	// adminID 可能是 string 或 uint64
	var id uint64
	switch v := adminID.(type) {
	case string:
		var err error
		id, err = parseUintID(v)
		if err != nil {
			response.BadRequest(c, "Invalid admin ID")
			return
		}
	case uint64:
		id = v
	case float64:
		id = uint64(v)
	default:
		response.BadRequest(c, "Invalid admin ID type")
		return
	}

	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.service.ChangePassword(id, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
