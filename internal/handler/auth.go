package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type AuthHandler struct {
	collectionService *service.CollectionService
	authService       *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		collectionService: service.NewCollectionService(),
		authService:       service.NewAuthService(),
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求体
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 注册用户
	result, err := h.authService.Register(collection, data)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// AuthWithPassword 邮箱密码登录
func (h *AuthHandler) AuthWithPassword(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req service.AuthWithPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 执行登录
	result, err := h.authService.AuthWithPassword(collection, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// AuthRefresh 刷新Token
func (h *AuthHandler) AuthRefresh(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req service.AuthRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 刷新Token
	result, err := h.authService.RefreshToken(collection, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// RequestPasswordReset 请求密码重置
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 发送密码重置邮件
	response.Success(c, gin.H{"message": "Password reset email sent"})
}

// ConfirmPasswordReset 确认密码重置
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req service.ConfirmPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 确认密码重置
	err = h.authService.ConfirmPasswordReset(collection, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Password reset successful"})
}

// RequestVerification 请求邮箱验证
func (h *AuthHandler) RequestVerification(c *gin.Context) {
	response.Success(c, gin.H{"message": "Verification email sent"})
}

// ConfirmVerification 确认邮箱验证
func (h *AuthHandler) ConfirmVerification(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 确认邮箱验证
	err = h.authService.ConfirmVerification(collection, req.Email)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Email verified successfully"})
}

// RequestEmailChange 请求邮箱变更
func (h *AuthHandler) RequestEmailChange(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, _ := c.Get("user_id")
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case string:
		fmt.Sscanf(v, "%d", &uid)
	}

	// 解析请求
	var req service.RequestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 请求邮箱变更
	err = h.authService.RequestEmailChange(collection, uid, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Verification email sent to new address"})
}

// ConfirmEmailChange 确认邮箱变更
func (h *AuthHandler) ConfirmEmailChange(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, _ := c.Get("user_id")
	var uid uint64
	switch v := userID.(type) {
	case uint64:
		uid = v
	case float64:
		uid = uint64(v)
	case string:
		fmt.Sscanf(v, "%d", &uid)
	}

	// 解析请求
	var req service.ConfirmEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 确认邮箱变更
	err = h.authService.ConfirmEmailChange(collection, uid, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Email changed successfully"})
}

// RequestOTP 请求验证码
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req service.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 请求验证码
	err = h.authService.RequestOTP(collection, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "验证码已发送"})
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	collectionName := c.Param("collection")

	// 获取集合信息
	collection, err := h.collectionService.GetByName(collectionName)
	if err != nil {
		response.NotFound(c, "Collection not found")
		return
	}

	// 验证是否为 Auth 类型
	if collection.Type != "auth" {
		response.BadRequest(c, "Collection is not an auth collection")
		return
	}

	// 解析请求
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 重置密码
	err = h.authService.ResetPassword(collection, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "密码重置成功"})
}
