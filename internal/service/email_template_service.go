package service

import (
	"sync"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"gorm.io/gorm"
)

// EmailTemplateService 邮件模板服务
type EmailTemplateService struct {
	db *gorm.DB
}

var emailTemplateService *EmailTemplateService
var emailTemplateOnce sync.Once

func NewEmailTemplateService() *EmailTemplateService {
	emailTemplateOnce.Do(func() {
		emailTemplateService = &EmailTemplateService{
			db: database.GetDB(),
		}
	})
	return emailTemplateService
}

// GetByType 根据类型获取模板
func (s *EmailTemplateService) GetByType(templateType string) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := s.db.Where("type = ?", templateType).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// List 获取所有模板
func (s *EmailTemplateService) List() ([]model.EmailTemplate, error) {
	var templates []model.EmailTemplate
	err := s.db.Order("type ASC").Find(&templates).Error
	return templates, err
}

// Create 创建模板
func (s *EmailTemplateService) Create(template *model.EmailTemplate) error {
	return s.db.Create(template).Error
}

// Update 更新模板
func (s *EmailTemplateService) Update(templateType string, updates map[string]interface{}) error {
	return s.db.Model(&model.EmailTemplate{}).Where("type = ?", templateType).Updates(updates).Error
}

// Delete 删除模板
func (s *EmailTemplateService) Delete(templateType string) error {
	return s.db.Where("type = ?", templateType).Delete(&model.EmailTemplate{}).Error
}

// InitDefaultTemplates 初始化默认模板
func (s *EmailTemplateService) InitDefaultTemplates() error {
	defaults := []model.EmailTemplate{
		{
			Type:    "verification",
			Subject: "验证您的邮箱",
			Body:    `<h1>验证您的邮箱</h1><p>您好，</p><p>请点击以下链接验证您的邮箱地址：</p><p><a href="{{{.VerificationURL}}}">验证邮箱</a></p><p>或者复制以下链接到浏览器：</p><p>{{{.VerificationURL}}}</p><p>此链接将在24小时后过期。</p>`,
			Enabled: true,
		},
		{
			Type:    "password-reset",
			Subject: "重置您的密码",
			Body:    `<h1>重置密码</h1><p>您好，</p><p>请点击以下链接重置您的密码：</p><p><a href="{{{.ResetURL}}}">重置密码</a></p><p>或者复制以下链接到浏览器：</p><p>{{{.ResetURL}}}</p><p>此链接将在1小时后过期。</p>`,
			Enabled: true,
		},
		{
			Type:    "email-change",
			Subject: "确认邮箱变更",
			Body:    `<h1>邮箱变更确认</h1><p>您好，</p><p>您请求将邮箱变更为：{{{ .NewEmail }}}</p><p>请点击以下链接确认：</p><p><a href="{{{.ConfirmURL}}}">确认变更</a></p><p>或者复制以下链接到浏览器：</p><p>{{{.ConfirmURL}}}</p><p>此链接将在1小时后过期。</p>`,
			Enabled: true,
		},
		{
			Type:    "welcome",
			Subject: "欢迎加入",
			Body:    `<h1>欢迎加入</h1><p>您好 {{{.Email}}}，</p><p>感谢您注册我们的服务！</p><p>如果您有任何问题，请随时联系我们。</p>`,
			Enabled: true,
		},
	}

	for _, t := range defaults {
		var existing model.EmailTemplate
		if err := s.db.Where("type = ?", t.Type).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := s.db.Create(&t).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
