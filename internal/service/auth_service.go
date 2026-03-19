package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/otp"
	"github.com/lvtao/go-gin-api-admin/pkg/password"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: database.GetDB(),
	}
}

// AuthWithPasswordRequest 密码登录请求
type AuthWithPasswordRequest struct {
	Identity string `json:"identity" binding:"required"` // 邮箱
	Password string `json:"password" binding:"required"`
}

// AuthWithPasswordResult 密码登录结果
type AuthWithPasswordResult struct {
	Token     string                 `json:"token"`
	Record    map[string]interface{} `json:"record"`
	RefreshToken string              `json:"refreshToken,omitempty"`
}

// AuthRefreshRequest 刷新Token请求
type AuthRefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// AuthRefreshResult 刷新Token结果
type AuthRefreshResult struct {
	Token string `json:"token"`
}

// ConfirmPasswordResetRequest 确认密码重置请求
type ConfirmPasswordResetRequest struct {
	Password string `json:"password" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// AuthWithPassword 密码登录
// identity 自动判断：带@为邮箱，手机号为mobile，其他为account
func (s *AuthService) AuthWithPassword(collection *model.Collection, req *AuthWithPasswordRequest) (*AuthWithPasswordResult, error) {
	tableName := collection.Name

	// 判断 identity 类型
	identityField := s.getIdentityField(req.Identity)

	// 查找用户
	var user map[string]interface{}
	if err := s.db.Table(tableName).Where(identityField+" = ?", req.Identity).Scan(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("账号或密码错误")
	}

	// 验证密码
	hashedPwd, ok := user["password"].(string)
	if !ok {
		return nil, errors.New("账号或密码错误")
	}

	if !password.VerifyPassword(hashedPwd, req.Password) {
		return nil, errors.New("账号或密码错误")
	}

	// 检查是否已验证（可选）
	verified, _ := user["verified"].(bool)
	if !verified {
		// 可以选择允许未验证用户登录，或返回错误
		// return nil, errors.New("账号未验证")
	}

	// 获取用户ID（支持uint和float64类型，因为JSON数字解析为float64）
	var userIDStr string
	switch v := user["id"].(type) {
	case uint:
		userIDStr = fmt.Sprintf("%d", v)
	case uint32:
		userIDStr = fmt.Sprintf("%d", v)
	case uint64:
		userIDStr = fmt.Sprintf("%d", v)
	case float64:
		userIDStr = fmt.Sprintf("%.0f", v)
	case string:
		userIDStr = v
	}

	// 获取邮箱用于生成 token
	email := ""
	if e, ok := user["email"].(string); ok {
		email = e
	} else if m, ok := user["mobile"].(string); ok {
		email = m
	} else if a, ok := user["account"].(string); ok {
		email = a
	}

	// 生成Token
	token, err := auth.GenerateUserToken(userIDStr, email, tableName)
	if err != nil {
		return nil, err
	}

	// 生成RefreshToken
	refreshToken, err := auth.GenerateRefreshTokenWithInfo(email, tableName)
	if err != nil {
		return nil, err
	}

	// 移除敏感信息
	delete(user, "password")
	delete(user, "tokenKey")

	// 更新用户的tokenKey
	tokenKey := generateTokenKey()
	s.db.Table(tableName).Where(identityField+" = ?", req.Identity).Update("tokenKey", tokenKey)

	return &AuthWithPasswordResult{
		Token:        token,
		Record:       user,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(collection *model.Collection, req *AuthRefreshRequest) (*AuthRefreshResult, error) {
	// 验证RefreshToken
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("Invalid refresh token")
	}

	// 生成新的AccessToken
	email := claims["email"].(string)
	collectionName := claims["collection"].(string)

	// 验证collection名称匹配
	if collectionName != collection.Name {
		return nil, errors.New("Invalid token")
	}

	// 获取用户ID
	var userIDStr string
	if userID, ok := claims["user_id"]; ok {
		userIDStr = fmt.Sprintf("%v", userID)
	}

	token, err := auth.GenerateUserToken(userIDStr, email, collectionName)
	if err != nil {
		return nil, err
	}

	return &AuthRefreshResult{
		Token: token,
	}, nil
}

// ConfirmPasswordReset 确认密码重置
func (s *AuthService) ConfirmPasswordReset(collection *model.Collection, req *ConfirmPasswordResetRequest) error {
	// TODO: 验证token并重置密码
	// 暂时返回成功
	return nil
}

// ConfirmVerification 确认邮箱验证
func (s *AuthService) ConfirmVerification(collection *model.Collection, email string) error {
	tableName := collection.Name

	// 查找用户
	var user map[string]interface{}
	if err := s.db.Table(tableName).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("User not found")
		}
		return err
	}

	// 更新验证状态
	return s.db.Table(tableName).Where("email = ?", email).Update("verified", true).Error
}

// Register 用户注册
// 支持验证码验证：如果提供了 code 参数，会先验证验证码
// identity 字段自动判断：带@为邮箱，手机号为mobile，其他为account
func (s *AuthService) Register(collection *model.Collection, data map[string]interface{}) (*map[string]interface{}, error) {
	tableName := collection.Name

	// 获取 identity（支持 email/identity 两种参数名）
	identity, ok := data["identity"].(string)
	if !ok {
		identity, ok = data["email"].(string)
		if !ok {
			return nil, errors.New("账号不能为空")
		}
	}

	pwd, ok := data["password"].(string)
	if !ok {
		return nil, errors.New("密码不能为空")
	}

	// 判断 identity 类型并获取对应字段名
	identityField := s.getIdentityField(identity)

	// 检查用户是否已存在
	var count int64
	if err := s.db.Table(tableName).Where(identityField+" = ?", identity).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该账号已被注册")
	}

	// 如果提供了验证码，验证它
	if code, ok := data["code"].(string); ok && code != "" {
		valid, err := otp.Verify(identity, "register", code)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, errors.New("验证码无效")
		}
	}

	// 加密密码
	hashedPassword, err := password.HashPassword(pwd)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 构建用户数据（ID由数据库自动生成）
	user := map[string]interface{}{
		"password": hashedPassword,
		"tokenKey": generateTokenKey(),
		"created":  now,
		"updated":  now,
	}

	// 设置 identity 字段
	user[identityField] = identity

	// 从集合字段定义中获取系统字段的默认值
	fieldDefaults := make(map[string]interface{})
	for _, field := range collection.Fields {
		switch field.Name {
		case "emailVisibility", "verified":
			if field.DefaultValue != nil {
				fieldDefaults[field.Name] = field.DefaultValue
			} else {
				// 如果没有设置默认值，使用 false
				fieldDefaults[field.Name] = false
			}
		}
	}
	user["emailVisibility"] = fieldDefaults["emailVisibility"]
	// 如果验证了，设置 verified 为 true
	if _, ok := data["code"].(string); ok && data["code"] != "" {
		user["verified"] = true
	} else {
		user["verified"] = fieldDefaults["verified"]
	}

	// 添加其他字段
	for _, field := range collection.Fields {
		if field.Name != "email" && field.Name != "mobile" && field.Name != "account" &&
			field.Name != "password" && field.Name != "emailVisibility" &&
			field.Name != "verified" && field.Name != "tokenKey" && field.Name != "code" &&
			field.Name != "identity" {
			if value, ok := data[field.Name]; ok {
				user[field.Name] = value
			} else if field.DefaultValue != nil {
				// 使用字段默认值
				user[field.Name] = field.DefaultValue
			}
		}
	}

	// 创建用户
	if err := s.db.Table(tableName).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 移除敏感信息
	delete(user, "password")
	delete(user, "tokenKey")

	return &user, nil
}

// RequestEmailChangeRequest 请求邮箱变更请求
type RequestEmailChangeRequest struct {
	NewEmail string `json:"newEmail" binding:"required"`
}

// RequestEmailChangeResult 请求邮箱变更结果
type RequestEmailChangeResult struct {
	ExpiresAt time.Time `json:"expiresAt"`
}

// ConfirmEmailChangeRequest 确认邮箱变更请求
type ConfirmEmailChangeRequest struct {
	NewEmail string `json:"newEmail" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// RequestEmailChange 请求邮箱变更
func (s *AuthService) RequestEmailChange(collection *model.Collection, userID uint, req *RequestEmailChangeRequest) error {
	tableName := collection.Name

	// 检查新邮箱是否已被使用
	var count int64
	if err := s.db.Table(tableName).Where("email = ?", req.NewEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("Email already in use")
	}

	// TODO: 发送验证邮件到新邮箱

	return nil
}

// ConfirmEmailChange 确认邮箱变更
func (s *AuthService) ConfirmEmailChange(collection *model.Collection, userID uint64, req *ConfirmEmailChangeRequest) error {
	tableName := collection.Name

	// TODO: 验证token

	// 检查邮箱是否已被使用
	var count int64
	if err := s.db.Table(tableName).Where("email = ?", req.NewEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("Email already in use")
	}

	// 更新邮箱
	if err := s.db.Table(tableName).Where("id = ?", userID).Update("email", req.NewEmail).Error; err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

// RequestOTPRequest 请求验证码请求
type RequestOTPRequest struct {
	Identity string `json:"identity" binding:"required"` // 邮箱/手机号/账号
	Type     string `json:"type" binding:"required"`    // register, password-reset, email-login
}

// RequestOTP 请求验证码
// 注册时：如果用户已存在，不允许发送验证码
// 找回密码/邮件登录：如果用户不存在，不允许发送验证码
func (s *AuthService) RequestOTP(collection *model.Collection, req *RequestOTPRequest) error {
	tableName := collection.Name

	// 判断 identity 类型
	identityField := s.getIdentityField(req.Identity)

	// 检查用户是否存在
	var count int64
	if err := s.db.Table(tableName).Where(identityField+" = ?", req.Identity).Count(&count).Error; err != nil {
		return err
	}

	switch req.Type {
	case "register":
		// 注册时，用户不能存在
		if count > 0 {
			return errors.New("该账号已被注册")
		}
	case "password-reset", "email-login":
		// 找回密码或邮件登录时，用户必须存在
		if count == 0 {
			return errors.New("该账号未注册")
		}
	default:
		return errors.New("无效的请求类型")
	}

	// 生成验证码
	code := otp.Generate(req.Identity, req.Type)

	// TODO: 发送邮件或短信
	// 这里暂时只打印验证码，实际应该发送邮件或短信
	fmt.Printf("[OTP] 发送验证码到 %s: %s (类型: %s, 字段: %s)\n", req.Identity, code, req.Type, identityField)

	return nil
}

// getIdentityField 根据 identity 值判断使用哪个字段
func (s *AuthService) getIdentityField(identity string) string {
	// 判断是否包含 @，包含则为邮箱
	if strings.Contains(identity, "@") {
		return "email"
	}
	// 判断是否为手机号（以1开头，11位数字）
	if matched, _ := regexp.MatchString(`^1\d{10}$`, identity); matched {
		return "mobile"
	}
	// 其他情况为账号
	return "account"
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Identity string `json:"identity" binding:"required"` // 邮箱/手机号/账号
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(collection *model.Collection, req *ResetPasswordRequest) error {
	tableName := collection.Name

	// 验证验证码
	valid, err := otp.Verify(req.Identity, "password-reset", req.Code)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("验证码无效")
	}

	// 判断 identity 类型
	identityField := s.getIdentityField(req.Identity)

	// 检查用户是否存在
	var count int64
	if err := s.db.Table(tableName).Where(identityField+" = ?", req.Identity).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("用户不存在")
	}

	// 加密新密码
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.db.Table(tableName).Where(identityField+" = ?", req.Identity).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
