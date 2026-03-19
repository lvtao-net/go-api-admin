package model

// OperationLog 操作日志
type OperationLog struct {
	BaseModel
	// 操作信息
	Collection string `gorm:"size:100;index" json:"collection,omitempty"` // 操作的集合
	RecordID   uint64 `gorm:"index" json:"recordId,omitempty"`           // 操作的记录ID
	Action     string `gorm:"size:20;not null;index" json:"action"`      // 操作类型: create, update, delete, login, logout
	// 用户信息
	UserID    uint64 `gorm:"index" json:"userId,omitempty"`       // 用户ID
	UserEmail string `gorm:"size:255" json:"userEmail,omitempty"` // 用户邮箱
	UserType  string `gorm:"size:10;index" json:"userType,omitempty"` // 用户类型: admin, user
	// 请求信息
	IP        string `gorm:"size:45" json:"ip,omitempty"`         // IP地址
	UserAgent string `gorm:"size:500" json:"userAgent,omitempty"` // User-Agent
	Method    string `gorm:"size:10" json:"method,omitempty"`     // 请求方法
	Path      string `gorm:"size:500" json:"path,omitempty"`      // 请求路径
	// 数据
	Request  string `gorm:"type:json" json:"request,omitempty"`  // 请求数据
	Response string `gorm:"type:json" json:"response,omitempty"` // 响应数据
	Status   int    `gorm:"default:200;index" json:"status"`     // 响应状态码
	// 消息
	Message string `gorm:"size:500" json:"message,omitempty"` // 日志消息
}

func (OperationLog) TableName() string {
	return SystemTablePrefix + "logs"
}
