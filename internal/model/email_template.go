package model

// EmailTemplate 邮件模板
type EmailTemplate struct {
	BaseModel
	Type    string `gorm:"size:50;uniqueIndex;not null" json:"type"` // 模板类型: verification, password-reset, email-change
	Subject string `gorm:"size:255;not null" json:"subject"`         // 邮件主题
	Body    string `gorm:"type:text;not null" json:"body"`           // 邮件正文(HTML)
	Enabled bool   `gorm:"default:true" json:"enabled"`              // 是否启用
}

func (EmailTemplate) TableName() string {
	return SystemTablePrefix + "email_templates"
}
