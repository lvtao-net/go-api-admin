package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/password"
	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService() *AdminService {
	return &AdminService{
		db: database.GetDB(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"identity" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResult 登录结果
type LoginResult struct {
	Token string     `json:"token"`
	Admin *AdminInfo `json:"admin"`
}

// AdminInfo 管理员信息
type AdminInfo struct {
	ID      uint64    `json:"id"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

// Login 管理员登录
func (s *AdminService) Login(req *LoginRequest) (*LoginResult, error) {
	var admin model.Admin

	// 查找管理员
	if err := s.db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// 验证密码
	if !password.VerifyPassword(admin.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// 生成Token（ID 转为字符串）
	token, err := auth.GenerateAdminToken(fmt.Sprintf("%d", admin.ID), admin.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		Admin: &AdminInfo{
			ID:      admin.ID,
			Email:   admin.Email,
			Created: admin.Created,
		},
	}, nil
}

// GetByID 通过ID获取管理员
func (s *AdminService) GetByID(id uint64) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// CreateAdmin 创建管理员
func (s *AdminService) CreateAdmin(email, pwd string) (*model.Admin, error) {
	// 检查邮箱是否已存在
	var count int64
	s.db.Model(&model.Admin{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := password.HashPassword(pwd)
	if err != nil {
		return nil, err
	}

	admin := &model.Admin{
		Email:    email,
		Password: hashedPassword,
		TokenKey: generateID() + generateID(),
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, err
	}

	return admin, nil
}

// UpdatePassword 更新密码
func (s *AdminService) UpdatePassword(id uint64, newPassword string) error {
	hashedPassword, err := password.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.Model(&model.Admin{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// ChangePassword 修改密码（需验证旧密码）
func (s *AdminService) ChangePassword(id uint64, oldPassword, newPassword string) error {
	// 获取管理员信息
	admin, err := s.GetByID(id)
	if err != nil {
		return errors.New("admin not found")
	}

	// 验证旧密码
	if !password.VerifyPassword(admin.Password, oldPassword) {
		return errors.New("invalid old password")
	}

	// 更新密码
	return s.UpdatePassword(id, newPassword)
}

// EnsureDefaultAdmin 确保存在默认管理员（包级别函数）
func EnsureDefaultAdmin() error {
	svc := NewAdminService()
	return svc.EnsureDefaultAdmin()
}

// EnsureDefaultAdmin 确保存在默认管理员
func (s *AdminService) EnsureDefaultAdmin() error {
	var count int64
	s.db.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 从配置文件读取管理员账号密码
	email := config.GlobalConfig.Admin.Email
	pwd := config.GlobalConfig.Admin.Password

	// 如果配置为空，使用默认值
	if email == "" {
		email = "admin@example.com"
	}
	if pwd == "" {
		pwd = "admin123456"
	}

	// 加密密码
	hashedPassword, err := password.HashPassword(pwd)
	if err != nil {
		return err
	}

	// 创建默认管理员（ID 自动递增）
	admin := &model.Admin{
		Email:    email,
		Password: hashedPassword,
		TokenKey: generateID() + generateID(),
	}

	return s.db.Create(admin).Error
}

// List 获取管理员列表
func (s *AdminService) List(page, perPage int) ([]model.Admin, int64, error) {
	var admins []model.Admin
	var total int64

	query := s.db.Model(&model.Admin{})
	query.Count(&total)

	offset := (page - 1) * perPage
	if err := query.Offset(offset).Limit(perPage).Order("created DESC").Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// Delete 删除管理员
func (s *AdminService) Delete(id uint64) error {
	// 检查是否是最后一个管理员
	var count int64
	s.db.Model(&model.Admin{}).Count(&count)
	if count <= 1 {
		return errors.New("cannot delete the last admin")
	}

	// 检查管理员是否存在
	var admin model.Admin
	if err := s.db.First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("admin not found")
		}
		return err
	}

	return s.db.Delete(&admin).Error
}

// Update 更新管理员
func (s *AdminService) Update(id uint64, email string) error {
	return s.db.Model(&model.Admin{}).Where("id = ?", id).Update("email", email).Error
}
