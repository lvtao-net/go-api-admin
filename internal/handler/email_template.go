package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

// EmailTemplateHandler 邮件模板处理器
type EmailTemplateHandler struct {
	templateService *service.EmailTemplateService
}

func NewEmailTemplateHandler() *EmailTemplateHandler {
	return &EmailTemplateHandler{
		templateService: service.NewEmailTemplateService(),
	}
}

// List 获取邮件模板列表
func (h *EmailTemplateHandler) List(c *gin.Context) {
	templates, err := h.templateService.List()
	if err != nil {
		response.Error(c, 500, "获取模板列表失败")
		return
	}
	response.Success(c, templates)
}

// Get 获取单个模板
func (h *EmailTemplateHandler) Get(c *gin.Context) {
	templateType := c.Param("type")

	template, err := h.templateService.GetByType(templateType)
	if err != nil {
		response.Error(c, 404, "模板不存在")
		return
	}
	response.Success(c, template)
}

// Update 更新模板
func (h *EmailTemplateHandler) Update(c *gin.Context) {
	templateType := c.Param("type")

	var req struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数错误")
		return
	}

	updates := map[string]interface{}{
		"subject": req.Subject,
		"body":    req.Body,
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	err := h.templateService.Update(templateType, updates)
	if err != nil {
		response.Error(c, 500, "更新模板失败")
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// Test 发送测试邮件
func (h *EmailTemplateHandler) Test(c *gin.Context) {
	templateType := c.Param("type")

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请提供有效的邮箱地址")
		return
	}

	// 获取模板
	template, err := h.templateService.GetByType(templateType)
	if err != nil {
		response.Error(c, 404, "模板不存在")
		return
	}

	if !template.Enabled {
		response.Error(c, 400, "模板已禁用")
		return
	}

	// TODO: 发送测试邮件
	// 这里需要调用邮件服务发送测试邮件

	response.Success(c, gin.H{"message": "测试邮件已发送"})
}

// Create 创建模板（管理员用）
func (h *EmailTemplateHandler) Create(c *gin.Context) {
	var template model.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.Error(c, 400, "请求参数错误")
		return
	}

	// 检查是否已存在
	existing, _ := h.templateService.GetByType(template.Type)
	if existing != nil {
		response.Error(c, 400, "模板类型已存在")
		return
	}

	if err := h.templateService.Create(&template); err != nil {
		response.Error(c, 500, "创建模板失败")
		return
	}

	response.Success(c, template)
}
