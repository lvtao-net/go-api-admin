package model

// Admin 管理员模型
type Admin struct {
	BaseModel
	Email    string `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	TokenKey string `gorm:"size:50;not null" json:"-"`
}

func (Admin) TableName() string {
	return SystemTablePrefix + "admins"
}
