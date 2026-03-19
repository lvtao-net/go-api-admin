package model

// Dictionary 字典表
type Dictionary struct {
	BaseModel
	Name        string           `gorm:"size:100;uniqueIndex;not null" json:"name"` // 字典名称
	Label       string           `gorm:"size:100" json:"label"`                     // 显示名称
	Description string           `gorm:"size:500" json:"description"`               // 描述
	System      bool             `gorm:"default:false" json:"system"`               // 系统字典（不可删除）
	Items       []DictionaryItem `gorm:"serializer:json" json:"items,omitempty"`    // 字典项
}

func (Dictionary) TableName() string {
	return SystemTablePrefix + "dictionaries"
}

// DictionaryItem 字典项
type DictionaryItem struct {
	ID          uint64 `json:"id"`                    // 字典项ID
	Label       string `json:"label"`                 // 显示名称
	Value       string `json:"value"`                 // 值
	Sort        int    `json:"sort"`                  // 排序
	Disabled    bool   `json:"disabled,omitempty"`    // 是否禁用
	Description string `json:"description,omitempty"` // 描述
	Color       string `json:"color,omitempty"`       // 颜色
}
