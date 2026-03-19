package model

// Collection 集合模型
type Collection struct {
	BaseModel
	Name        string            `gorm:"size:255;uniqueIndex;not null" json:"name"`
	Type        string            `gorm:"size:20;not null;default:'base'" json:"type"` // base, auth, view, transaction
	System      bool              `gorm:"default:false" json:"system"`
	Fields      []CollectionField `gorm:"serializer:json" json:"fields"`
	Indexes     []CollectionIndex `gorm:"serializer:json" json:"indexes,omitempty"`
	// API 规则
	ListRule   *string `json:"listRule,omitempty"`
	ViewRule   *string `json:"viewRule,omitempty"`
	CreateRule *string `json:"createRule,omitempty"`
	UpdateRule *string `json:"updateRule,omitempty"`
	DeleteRule *string `json:"deleteRule,omitempty"`
	// 更新限制（集合级别）
	UpdateFields []string `gorm:"serializer:json" json:"updateFields,omitempty"`
	// 主键字段配置（默认为 id）
	PrimaryKeyField string `gorm:"size:50;default:id" json:"primaryKeyField,omitempty"`
	// 可查找字段配置（支持通过这些字段查询单条记录）
	LookupFields []LookupFieldConfig `gorm:"serializer:json" json:"lookupFields,omitempty"`
	// 自定义路由参数（用于视图/事务集合）
	RouteParams []RouteParam `gorm:"serializer:json" json:"routeParams,omitempty"`
	// 视图集合
	ViewQuery   string   `gorm:"type:text" json:"viewQuery,omitempty"`          // 视图查询 SQL
	ViewRelated []string `gorm:"serializer:json" json:"viewRelated,omitempty"` // 视图关联的集合
	// 事务集合
	TransactionSteps []TransactionStep `gorm:"serializer:json" json:"transactionSteps,omitempty"` // 事务步骤
	// UI/UX扩展
	Label       string `gorm:"size:100" json:"label,omitempty"`        // 显示名称（别名）
	Description string `gorm:"size:500" json:"description,omitempty"` // 描述
	Icon        string `gorm:"size:50" json:"icon,omitempty"`         // 菜单图标
	MenuHidden  bool   `gorm:"default:false" json:"menuHidden"`       // 是否在菜单隐藏
	Sort        int    `gorm:"default:0" json:"sort"`                 // 排序
}

// LookupFieldConfig 可查找字段配置
type LookupFieldConfig struct {
	Field      string `json:"field"`                // 字段名，如 uid, order_no, slug
	Required   bool   `json:"required,omitempty"`   // 是否必填
	Validation string `json:"validation,omitempty"` // 验证规则
}

// RouteParam 路由参数配置
type RouteParam struct {
	Name        string `json:"name"`                  // 参数名
	Type        string `json:"type"`                  // 类型：string, number, bool
	Source      string `json:"source"`                // 来源：path, query, body
	Required    bool   `json:"required,omitempty"`    // 是否必填
	Default     string `json:"default,omitempty"`     // 默认值
	Description string `json:"description,omitempty"` // 说明
}

// TransactionStep 事务步骤
type TransactionStep struct {
	// 步骤基本信息
	Name        string `json:"name"`                  // 步骤名称（用于显示）
	Description string `json:"description,omitempty"` // 步骤描述

	// 步骤类型: query, validate, update, insert, delete
	Type string `json:"type"`

	// 操作的表名（集合名称）
	Table string `json:"table,omitempty"`

	// 查询结果别名（用于后续步骤引用）
	Alias string `json:"alias,omitempty"`

	// 查询/更新/删除条件
	Conditions []TransactionCondition `json:"conditions,omitempty"`

	// 插入/更新的数据
	Data map[string]interface{} `json:"data,omitempty"`

	// 验证条件（validate 类型）
	ValidateCondition string `json:"validateCondition,omitempty"`

	// 是否必须存在（query 类型）
	Required bool `json:"required,omitempty"`

	// 错误消息
	Error string `json:"error,omitempty"`

	// 失败时的处理: fail, skip
	OnError string `json:"onError,omitempty"`
}

// TransactionCondition 事务条件
type TransactionCondition struct {
	// 字段名
	Field string `json:"field"`

	// 操作符: =, !=, >, <, >=, <=, in, like
	Operator string `json:"operator,omitempty"`

	// 值（固定值）
	Value interface{} `json:"value,omitempty"`

	// 从上下文获取值: params.xxx, user.id, alias.field
	// 例如: params.orderId, user.id, order.totalAmount
	ValueFrom string `json:"valueFrom,omitempty"`
}

func (Collection) TableName() string {
	return SystemTablePrefix + "collections"
}

// CollectionField 字段定义
type CollectionField struct {
	Name string `json:"name"`
	// 字段类型: text, number, bool, email, url, date, datetime, select, radio, 
	// checkbox, relation, file, image, editor, json, password, tel, textarea
	Type string `json:"type"`
	// 展示类型: input, textarea, rich_text, upload, image, select, radio, 
	// checkbox, date, datetime, switch, slider, color, rate, password, json
	// 如果不设置，默认根据 type 推断
	DisplayType string `json:"displayType,omitempty"`
	// 字段选项（基础）
	Required     bool        `json:"required"`
	Unique       bool        `json:"unique,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	// 显示标签和提示
	Label       string `json:"label,omitempty"`       // 字段显示名称
	Placeholder string `json:"placeholder,omitempty"` // 输入提示
	Description string `json:"description,omitempty"` // 字段描述
	// 验证选项
	Min      interface{} `json:"min,omitempty"`      // 最小值/最小长度
	Max      interface{} `json:"max,omitempty"`      // 最大值/最大长度
	Pattern  string      `json:"pattern,omitempty"`  // 正则验证
	MinValue interface{} `json:"minValue,omitempty"` // 数字最小值
	MaxValue interface{} `json:"maxValue,omitempty"` // 数字最大值
	// 内置验证规则: required, email, phone, url, idcard, ip, ipv4, ipv6, number,
	// integer, positive, negative, alpha, alphanum, chinese, date, datetime,
	// min_length, max_length, range_length, min_value, max_value, range_value,
	// password_strength, credit_card, wechat, qq, bank_card
	ValidationRules []string `json:"validationRules,omitempty"`
	// 验证错误消息（可选自定义）
	ValidationMessages map[string]string `json:"validationMessages,omitempty"`
	// UI 选项
	Width     string `json:"width,omitempty"`     // 字段宽度: small, medium, large, full
	Sort      int    `json:"sort,omitempty"`      // 排序
	Group     string `json:"group,omitempty"`     // 字段分组
	Readonly  bool   `json:"readonly,omitempty"`  // 只读
	Disabled  bool   `json:"disabled,omitempty"`  // 禁用
	Multiple  bool   `json:"multiple,omitempty"`  // 多选（用于select/file）
	Step      string `json:"step,omitempty"`      // 数字步进
	Rows      int    `json:"rows,omitempty"`      // textarea行数
	Accept    string `json:"accept,omitempty"`    // 文件上传类型限制
	MaxSize   int64  `json:"maxSize,omitempty"`   // 文件大小限制(字节)
	MaxCount  int    `json:"maxCount,omitempty"`  // 文件数量限制
	// 字段级权限（前端UI）
	Hidden       bool `json:"hidden,omitempty"`       // 是否隐藏（不可见）
	HiddenOnList bool `json:"hiddenOnList,omitempty"` // 列表页隐藏
	HiddenOnView bool `json:"hiddenOnView,omitempty"` // 详情页隐藏
	HiddenOnForm bool `json:"hiddenOnForm,omitempty"` // 表单隐藏
	Editable     bool `json:"editable,omitempty"`     // 是否可编辑（默认true）
	// API 级别权限控制
	APIDisabled   bool `json:"apiDisabled,omitempty"`   // API 禁止访问
	APIReadOnly   bool `json:"apiReadOnly,omitempty"`   // API 只读
	APIWriteOnly  bool `json:"apiWriteOnly,omitempty"`  // API 只写
	APIHiddenList bool `json:"apiHiddenList,omitempty"` // API 列表接口隐藏
	APIHiddenView bool `json:"apiHiddenView,omitempty"` // API 详情接口隐藏
	// 更新限制
	UpdateOnly bool `json:"updateOnly,omitempty"` // 仅允许更新
	// 关联关系配置（用于relation类型）
	RelationCollection string `json:"relationCollection,omitempty"` // 关联集合名称
	RelationField      string `json:"relationField,omitempty"`      // 关联字段（默认id）
	RelationLabelField string `json:"relationLabelField,omitempty"` // 显示字段（用于下拉选择）
	RelationType       string `json:"relationType,omitempty"`       // 关联类型: has_one, has_many, belongs_to, many_to_many
	RelationCascade    bool   `json:"relationCascade,omitempty"`    // 级联删除
	RelationMax        int    `json:"relationMax,omitempty"`        // 多对多最大数量
	RelationName       string `json:"relationName,omitempty"`       // 关联名称（用于反向查询）
	// 选择项配置（用于select/radio/checkbox）
	Options     map[string]interface{} `json:"options,omitempty"`     // 通用选项
	FieldOptions []FieldOption         `json:"fieldOptions,omitempty"` // 选项列表
	Dictionary   string                `json:"dictionary,omitempty"`   // 关联字典名称
	// 富文本编辑器配置
	EditorMode    string `json:"editorMode,omitempty"`    // 编辑器模式: simple, standard, full
	EditorToolbar string `json:"editorToolbar,omitempty"` // 自定义工具栏
	// 日期时间格式配置
	DateFormat string `json:"dateFormat,omitempty"` // 日期格式: date(年-月-日), datetime(年-月-日 时:分:秒), time(时:分:秒), 或自定义格式如 2006-01-02
	// 其他扩展
	Extra map[string]interface{} `json:"extra,omitempty"` // 自定义扩展配置
}

// FieldOption 字段选项
type FieldOption struct {
	Label    string `json:"label"`              // 显示名称
	Value    string `json:"value"`              // 值
	Disabled bool   `json:"disabled,omitempty"` // 是否禁用
	Sort     int    `json:"sort,omitempty"`     // 排序
	Color    string `json:"color,omitempty"`    // 颜色标识
}

// CollectionIndex 索引定义
type CollectionIndex struct {
	Name   string   `json:"name,omitempty"`
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}
