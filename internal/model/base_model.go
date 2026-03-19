package model

import (
	"time"

	"gorm.io/gorm"
)

// 系统表前缀
const SystemTablePrefix = "_"

// BaseModel 基础模型（所有表统一使用bigint自增ID）
// 所有集合表（包括内置表和用户创建的表）都强制使用此模型
type BaseModel struct {
	ID      uint64         `gorm:"primaryKey;autoIncrement" json:"id"` // bigint自增ID
	Created time.Time      `gorm:"autoCreateTime" json:"created"`      // 创建时间
	Updated time.Time      `gorm:"autoUpdateTime" json:"updated"`      // 更新时间
	Deleted gorm.DeletedAt `gorm:"index" json:"-"`                     // 软删除
}

// TableName 返回基础表名（由子类覆盖）
func (BaseModel) TableName() string {
	return ""
}
