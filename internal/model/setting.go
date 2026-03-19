package model

import (
	"time"

	"gorm.io/gorm"
)

// Setting 设置模型
type Setting struct {
	ID      uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Key     string         `gorm:"size:255;uniqueIndex;not null" json:"key"`
	Value   string         `gorm:"type:text" json:"value"`
	Updated time.Time      `gorm:"autoUpdateTime" json:"updated"`
	Deleted gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Setting) TableName() string {
	return SystemTablePrefix + "settings"
}
