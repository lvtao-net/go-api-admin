package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/repository"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
)

type CollectionService struct {
	repo *repository.CollectionRepository
}

func NewCollectionService() *CollectionService {
	return &CollectionService{
		repo: repository.NewCollectionRepository(database.GetDB()),
	}
}

// CreateCollectionRequest 创建集合请求
type CreateCollectionRequest struct {
	Name             string                    `json:"name" binding:"required"`
	Label            string                    `json:"label"`              // 中文别名
	Type             string                    `json:"type"`
	Description      string                    `json:"description"`        // 描述
	Fields           []model.CollectionField   `json:"fields"`
	ListRule         *string                   `json:"listRule"`
	ViewRule         *string                   `json:"viewRule"`
	CreateRule       *string                   `json:"createRule"`
	UpdateRule       *string                   `json:"updateRule"`
	DeleteRule       *string                   `json:"deleteRule"`
	ViewQuery        *string                   `json:"viewQuery,omitempty"`    // 视图查询 SQL
	ViewRelated      []string                  `json:"viewRelated,omitempty"`  // 视图关联的集合
	TransactionSteps []model.TransactionStep   `json:"transactionSteps,omitempty"` // 事务步骤（事务集合专用）
}

// UpdateCollectionRequest 更新集合请求
type UpdateCollectionRequest struct {
	Label             *string                   `json:"label"`
	Fields            []model.CollectionField   `json:"fields"`
	ListRule          *string                   `json:"listRule"`
	ViewRule          *string                   `json:"viewRule"`
	CreateRule        *string                   `json:"createRule"`
	UpdateRule        *string                   `json:"updateRule"`
	DeleteRule        *string                   `json:"deleteRule"`
	ViewQuery         *string                   `json:"viewQuery,omitempty"`           // 视图查询 SQL
	TransactionSteps  []model.TransactionStep   `json:"transactionSteps,omitempty"`  // 事务步骤
}

// Create 创建集合
func (s *CollectionService) Create(req *CreateCollectionRequest) (*model.Collection, error) {
	// 验证名称
	if err := validateCollectionName(req.Name); err != nil {
		return nil, err
	}

	// 检查名称是否已存在
	exists, err := s.repo.Exists(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}
	if exists {
		return nil, errors.New("collection with this name already exists")
	}

	// 设置默认类型
	if req.Type == "" {
		req.Type = "base"
	}

	// 验证类型
	if req.Type != "base" && req.Type != "auth" && req.Type != "view" && req.Type != "transaction" {
		return nil, errors.New("invalid collection type, must be 'base', 'auth', 'view' or 'transaction'")
	}

	// 视图集合必须有查询语句
	if req.Type == "view" && (req.ViewQuery == nil || *req.ViewQuery == "") {
		return nil, errors.New("view collection requires a viewQuery")
	}

	// 事务集合必须有事务步骤
	if req.Type == "transaction" && len(req.TransactionSteps) == 0 {
		return nil, errors.New("transaction collection requires transactionSteps")
	}

	// 创建集合对象（ID 自动递增）
	collection := &model.Collection{
		Name:             req.Name,
		Label:            req.Label,
		Type:             req.Type,
		Description:      req.Description,
		System:           false,
		Fields:           req.Fields,
		ListRule:         req.ListRule,
		ViewRule:         req.ViewRule,
		CreateRule:       req.CreateRule,
		UpdateRule:       req.UpdateRule,
		DeleteRule:       req.DeleteRule,
		TransactionSteps: req.TransactionSteps,
	}
	if req.ViewQuery != nil {
		collection.ViewQuery = *req.ViewQuery
	}
	collection.ViewRelated = req.ViewRelated

	// 如果是 Auth 类型，添加默认字段
	if req.Type == "auth" {
		collection.Fields = mergeAuthFields(collection.Fields)
	}

	// 保存到数据库
	if err := s.repo.Create(collection); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	// 动态创建表（非视图类型和非事务类型）
	if req.Type == "view" {
		// 创建视图
		if err := createView(collection); err != nil {
			_ = s.repo.Delete(collection.ID)
			return nil, fmt.Errorf("failed to create view: %w", err)
		}
	} else if req.Type != "transaction" {
		// 创建数据库表（基础集合和认证集合）
		if err := createTableForCollection(collection); err != nil {
			// 回滚：删除集合记录
			_ = s.repo.Delete(collection.ID)
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}
	// 事务集合不需要创建表或视图，只保存配置

	return collection, nil
}

// GetByID 通过ID获取集合（接受字符串ID）
func (s *CollectionService) GetByID(id string) (*model.Collection, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %s", id)
	}
	return s.repo.GetByID(uid)
}

// GetByName 通过名称获取集合
func (s *CollectionService) GetByName(name string) (*model.Collection, error) {
	return s.repo.GetByName(name)
}

// List 获取集合列表
func (s *CollectionService) List(page, perPage int) ([]model.Collection, int64, error) {
	return s.repo.List(page, perPage)
}

// Update 更新集合（接受字符串ID）
func (s *CollectionService) Update(id string, req *UpdateCollectionRequest) (*model.Collection, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %s", id)
	}
	collection, err := s.repo.GetByID(uid)
	if err != nil {
		return nil, err
	}

	// 更新标签
	if req.Label != nil {
		collection.Label = *req.Label
	}

	// 更新字段
	if req.Fields != nil {
		// 如果是 Auth 类型，保留系统字段
		if collection.Type == "auth" {
			collection.Fields = mergeAuthFields(req.Fields)
		} else {
			collection.Fields = req.Fields
		}
	}

	// 更新规则
	if req.ListRule != nil {
		collection.ListRule = req.ListRule
	}
	if req.ViewRule != nil {
		collection.ViewRule = req.ViewRule
	}
	if req.CreateRule != nil {
		collection.CreateRule = req.CreateRule
	}
	if req.UpdateRule != nil {
		collection.UpdateRule = req.UpdateRule
	}
	if req.DeleteRule != nil {
		collection.DeleteRule = req.DeleteRule
	}

	// 更新视图查询（视图集合）
	if req.ViewQuery != nil && collection.Type == "view" {
		collection.ViewQuery = *req.ViewQuery
		if err := createView(collection); err != nil {
			return nil, fmt.Errorf("failed to update view: %w", err)
		}
	}

	// 更新事务步骤（事务集合）
	if req.TransactionSteps != nil && collection.Type == "transaction" {
		collection.TransactionSteps = req.TransactionSteps
	}

	if err := s.repo.Update(collection); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	// 更新表结构（非视图和事务类型）
	if collection.Type != "view" && collection.Type != "transaction" {
		if err := updateTableForCollection(collection); err != nil {
			return nil, fmt.Errorf("failed to update table: %w", err)
		}
	}

	return collection, nil
}

// UpdateByName 通过名称更新集合
func (s *CollectionService) UpdateByName(name string, req *UpdateCollectionRequest) (*model.Collection, error) {
	collection, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	return s.Update(strconv.FormatUint(uint64(collection.ID), 10), req)
}

// Delete 删除集合（接受字符串ID）
func (s *CollectionService) Delete(id string) error {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid id: %s", id)
	}
	collection, err := s.repo.GetByID(uid)
	if err != nil {
		return err
	}

	// 不允许删除系统集合
	if collection.System {
		return errors.New("cannot delete system collection")
	}

	// 删除表
	if err := dropTableForCollection(collection); err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	// 删除记录
	return s.repo.Delete(uid)
}

// DeleteByName 通过名称删除集合
func (s *CollectionService) DeleteByName(name string) error {
	collection, err := s.repo.GetByName(name)
	if err != nil {
		return err
	}
	return s.Delete(strconv.FormatUint(uint64(collection.ID), 10))
}

// CheckDeleteCheckData 检查删除集合时的数据情况
type CheckDeleteData struct {
	HasData      bool  `json:"hasData"`      // 是否有数据
	RecordCount  int64 `json:"recordCount"`  // 记录数量
	CanDelete    bool  `json:"canDelete"`    // 是否可以删除
	RelatedCount int   `json:"relatedCount"` // 关联集合数量
	Related      []string `json:"related"`   // 关联的集合名称
}

// CheckDelete 检查删除集合时的数据
func (s *CollectionService) CheckDelete(name string) (*CheckDeleteData, error) {
	collection, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}

	result := &CheckDeleteData{
		CanDelete: true,
		Related:   []string{},
	}

	// 系统集合不允许删除
	if collection.System {
		result.CanDelete = false
		return result, nil
	}

	// 检查是否有数据
	db := database.GetDB()
	tableName := collection.Name
	var count int64
	err = db.Table(tableName).Count(&count).Error
	if err == nil {
		result.RecordCount = count
		result.HasData = count > 0
		if count > 0 {
			result.CanDelete = false
		}
	}

	// 检查是否有其他集合关联此集合
	allCollections, _, _ := s.repo.List(1, 1000)
	related := []string{}
	for _, col := range allCollections {
		if col.ID == collection.ID {
			continue
		}
		for _, field := range col.Fields {
			if field.Type == "relation" && field.RelationCollection == name {
				related = append(related, col.Name)
			}
		}
	}
	result.Related = related
	result.RelatedCount = len(related)
	if len(related) > 0 {
		result.CanDelete = false
	}

	return result, nil
}

// validateCollectionName 验证集合名称
func validateCollectionName(name string) error {
	if len(name) < 1 || len(name) > 255 {
		return errors.New("collection name must be between 1 and 255 characters")
	}
	// 不能以数字开头，只能包含字母、数字、下划线
	for i, r := range name {
		if i == 0 {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return errors.New("collection name must start with a letter")
			}
		} else {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
				return errors.New("collection name can only contain letters, numbers, and underscores")
			}
		}
	}
	// 不能是保留名称
	reserved := []string{"admins", "collections", "settings", "logs", "files", "realtime"}
	for _, r := range reserved {
		if name == r || name == "_"+r {
			return fmt.Errorf("collection name '%s' is reserved", name)
		}
	}
	return nil
}

// generateID 生成唯一ID
func generateID() string {
	id := uuid.New().String()
	// 取前15位
	return id[:8] + id[9:13] + id[14:15]
}

// mergeAuthFields 合并Auth默认字段
func mergeAuthFields(fields []model.CollectionField) []model.CollectionField {
	// 使用统一的系统字段默认配置
	systemFieldDefaults := getSystemFieldDefaults()

	// 构建字段映射，以用户提交的为准
	fieldMap := make(map[string]model.CollectionField)

	// 先添加用户提交的所有字段
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	// 确保系统字段存在，如果用户没有提交则使用默认值
	for name, defaultField := range systemFieldDefaults {
		if _, exists := fieldMap[name]; !exists {
			fieldMap[name] = defaultField
		} else {
			// 用户提交了该字段，但需要确保 type 和 required 不被修改
			existing := fieldMap[name]
			existing.Type = defaultField.Type
			existing.Required = defaultField.Required
			fieldMap[name] = existing
		}
	}

	// 按固定顺序返回字段
	result := []model.CollectionField{}
	for _, name := range []string{"email", "emailVisibility", "verified", "password", "tokenKey"} {
		if f, ok := fieldMap[name]; ok {
			result = append(result, f)
			delete(fieldMap, name)
		}
	}

	// 添加其他用户自定义字段
	for _, f := range fieldMap {
		result = append(result, f)
	}

	return result
}

// createTableForCollection 为集合创建数据库表
func createTableForCollection(collection *model.Collection) error {
	db := database.GetDB()
	tableName := collection.Name

	// 构建字段定义
	fieldDefs := buildFieldDefinitions(collection.Fields)
	
	// 构建建表SQL - 使用 BIGINT UNSIGNED AUTO_INCREMENT 作为主键
	sql := fmt.Sprintf("CREATE TABLE `%s` (", tableName)
	sql += "`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,"
	sql += "`created` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,"
	sql += "`updated` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,"
	sql += "`deleted_at` DATETIME(3) NULL DEFAULT NULL,"
	
	for _, fieldDef := range fieldDefs {
		sql += fmt.Sprintf("`%s` %s,", fieldDef.Name, fieldDef.Definition)
	}
	
	sql = sql[:len(sql)-1] // 移除最后的逗号
	sql += ", INDEX `idx_deleted_at` (`deleted_at`)"
	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"

	return db.Exec(sql).Error
}

// updateTableForCollection 更新集合的数据库表
func updateTableForCollection(collection *model.Collection) error {
	db := database.GetDB()
	tableName := collection.Name

	// 获取当前表结构
	var existingColumns []string
	rows, err := db.Raw(fmt.Sprintf("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '%s'", tableName)).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var col string
		rows.Scan(&col)
		existingColumns = append(existingColumns, col)
	}

	// 构建需要添加的字段
	fieldDefs := buildFieldDefinitions(collection.Fields)
	
	for _, fieldDef := range fieldDefs {
		found := false
		for _, col := range existingColumns {
			if col == fieldDef.Name {
				found = true
				break
			}
		}
		if !found && fieldDef.Name != "id" && fieldDef.Name != "created" && fieldDef.Name != "updated" && fieldDef.Name != "deleted_at" {
			sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", tableName, fieldDef.Name, fieldDef.Definition)
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		}
	}

	// 确保 deleted_at 字段存在（旧表迁移）
	deletedAtExists := false
	for _, col := range existingColumns {
		if col == "deleted_at" {
			deletedAtExists = true
			break
		}
	}
	if !deletedAtExists {
		sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `deleted_at` DATETIME(3) NULL DEFAULT NULL", tableName)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
		// 添加索引
		db.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `idx_deleted_at` (`deleted_at`)", tableName))
	}

	return nil
}

// dropTableForCollection 删除集合的数据库表
func dropTableForCollection(collection *model.Collection) error {
	db := database.GetDB()
	// 如果是视图，删除视图
	if collection.Type == "view" {
		return db.Exec(fmt.Sprintf("DROP VIEW IF EXISTS `%s`", collection.Name)).Error
	}
	return db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", collection.Name)).Error
}

// createView 创建视图
func createView(collection *model.Collection) error {
	db := database.GetDB()
	viewQuery := collection.ViewQuery
	if viewQuery == "" {
		return errors.New("view query is required")
	}

	// 创建视图 SQL
	sql := fmt.Sprintf("CREATE OR REPLACE VIEW `%s` AS %s", collection.Name, viewQuery)
	return db.Exec(sql).Error
}

// QueryView 查询视图数据
func (s *CollectionService) QueryView(collection *model.Collection, req *ListRecordsRequest) (*ListResult, error) {
	if collection.Type != "view" {
		return nil, errors.New("not a view collection")
	}

	viewQuery := collection.ViewQuery
	if viewQuery == "" {
		return nil, errors.New("view query is empty")
	}

	// 构建查询
	query := "SELECT * FROM (" + viewQuery + ") AS view_table"

	// 应用过滤条件（使用参数化查询）
	var filterArgs []interface{}
	if req.Filter != "" {
		conditions, args := parseFilter(req.Filter)
		filterArgs = args
		if conditions != "" {
			query += " WHERE " + conditions
		}
	}

	// 应用排序
	if req.Sort != "" {
		order := parseSort(req.Sort)
		query += " ORDER BY " + order
	} else {
		query += " ORDER BY created DESC"
	}

	// 统计总数
	var total int64
	db := database.GetDB()
	countQuery := "SELECT COUNT(*) FROM (" + viewQuery + ") AS view_table"
	if req.Filter != "" {
		conditions, args := parseFilter(req.Filter)
		if conditions != "" {
			countQuery += " WHERE " + conditions
		}
		// count查询使用相同的参数
		if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
			return nil, fmt.Errorf("failed to count view records: %w", err)
		}
	} else {
		if err := db.Raw(countQuery).Scan(&total).Error; err != nil {
			return nil, fmt.Errorf("failed to count view records: %w", err)
		}
	}

	// 分页查询
	offset := (req.Page - 1) * req.PerPage
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", req.PerPage, offset)

	var records []map[string]interface{}
	if err := db.Raw(query, filterArgs...).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query view: %w", err)
	}

	// 转换结果
	items := make([]RecordResult, 0, len(records))
	for _, record := range records {
		items = append(items, mapToRecordResult(record, nil))
	}

	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	return &ListResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

// FieldDefinition 字段定义
type FieldDefinition struct {
	Name       string
	Definition string
}

// buildFieldDefinitions 构建字段定义
func buildFieldDefinitions(fields []model.CollectionField) []FieldDefinition {
	var defs []FieldDefinition

	for _, field := range fields {
		def := buildFieldDefinition(field)
		defs = append(defs, def)
	}

	return defs
}

// buildFieldDefinition 构建单个字段定义
func buildFieldDefinition(field model.CollectionField) FieldDefinition {
	var def string

	switch field.Type {
	case "text", "password":
		def = "TEXT"
	case "number":
		def = "DOUBLE"
	case "bool":
		def = "BOOLEAN"
	case "email":
		def = "VARCHAR(255)"
	case "url":
		def = "VARCHAR(2048)"
	case "date":
		def = "DATETIME"
	case "select", "radio":
		// select/radio 类型存储单个值
		def = "VARCHAR(255)"
	case "checkbox":
		// checkbox 类型存储 JSON 数组（多选）
		def = "VARCHAR(500)"
	case "relation":
		// relation 字段存储关联记录的 ID
		// 如果是多对多，存储 JSON 数组
		if field.RelationMax > 1 || (field.Options != nil && field.Options["maxSelect"] != nil) {
			def = "VARCHAR(255)" // 存储 JSON 数组
		} else {
			def = "VARCHAR(15)" // 单个 ID
		}
	case "file":
		def = "VARCHAR(255)"
	case "editor":
		def = "LONGTEXT"
	case "json":
		def = "JSON"
	default:
		def = "TEXT"
	}

	if field.Required && field.Type != "bool" {
		def += " NOT NULL"
	}

	if field.DefaultValue != nil {
		if field.Type == "bool" {
			// MySQL 布尔类型默认值需要用 TRUE/FALSE 关键字
			// 处理各种可能的 bool 值类型
			var boolVal bool
			switch v := field.DefaultValue.(type) {
			case bool:
				boolVal = v
			case string:
				boolVal = v == "true" || v == "1"
			case float64:
				boolVal = v != 0
			case int:
				boolVal = v != 0
			default:
				boolVal = false
			}
			if boolVal {
				def += " DEFAULT TRUE"
			} else {
				def += " DEFAULT FALSE"
			}
		} else {
			def += fmt.Sprintf(" DEFAULT '%v'", field.DefaultValue)
		}
	}

	return FieldDefinition{
		Name:       field.Name,
		Definition: def,
	}
}

// InitSystemCollections 初始化系统集合
func InitSystemCollections() error {
	db := database.GetDB()

	// 检查是否已初始化
	var count int64
	db.Model(&model.Collection{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 系统字段默认配置
	systemFields := getSystemFieldDefaults()

	// 创建默认的 users Auth 集合（ID 自动递增）
	usersFields := []model.CollectionField{
		systemFields["email"],
		systemFields["emailVisibility"],
		systemFields["verified"],
		systemFields["password"],
		systemFields["tokenKey"],
		{
			Name:        "nickName",
			Type:        "text",
			Label:       "昵称",
			Description: "用户昵称",
			Required:    false,
		},
	}

	usersCollection := &model.Collection{
		Name:   "users",
		Type:   "auth",
		System: false,
		Label:  "用户",
		Fields: usersFields,
	}

	if err := db.Create(usersCollection).Error; err != nil {
		return err
	}

	// 创建对应的表
	return createTableForCollection(usersCollection)
}

// getSystemFieldDefaults 获取系统字段默认配置
func getSystemFieldDefaults() map[string]model.CollectionField {
	return map[string]model.CollectionField{
		"email": {
			Name:        "email",
			Type:        "email",
			Label:       "邮箱",
			Description: "用户邮箱地址，用于登录和接收通知",
			Required:    true,
			Unique:      true,
		},
		"emailVisibility": {
			Name:         "emailVisibility",
			Type:         "bool",
			Label:        "邮箱公开",
			Description:  "是否公开邮箱地址，公开后其他用户可以看到您的邮箱",
			Required:     false,
			DefaultValue: false,
		},
		"verified": {
			Name:         "verified",
			Type:         "bool",
			Label:        "已验证",
			Description:  "邮箱是否已验证",
			Required:     false,
			DefaultValue: false,
		},
		"password": {
			Name:        "password",
			Type:        "password",
			Label:       "密码",
			Description: "用户登录密码，至少8位字符",
			Required:    true,
		},
		"tokenKey": {
			Name:        "tokenKey",
			Type:        "text",
			Label:       "令牌密钥",
			Description: "用于生成认证令牌的随机密钥，系统自动生成",
			Required:    false,
		},
	}
}

// PreviewView 预览视图集合数据
func (s *CollectionService) PreviewView(name string, page, perPage int) (*ListResult, error) {
	collection, err := s.repo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("collection not found")
	}

	if collection.Type != "view" {
		return nil, fmt.Errorf("collection is not a view")
	}

	viewQuery := collection.ViewQuery
	if viewQuery == "" {
		return nil, fmt.Errorf("view query is empty")
	}

	if err := validateViewQuery(viewQuery); err != nil {
		return nil, fmt.Errorf("unsafe view query: %w", err)
	}

	db := database.GetDB()

	countQuery := "SELECT COUNT(*) FROM (" + viewQuery + ") AS view_table"
	var total int64
	if err := db.Raw(countQuery).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count view records: %w", err)
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf("SELECT * FROM (%s) AS view_table LIMIT %d OFFSET %d", viewQuery, perPage, offset)

	var records []map[string]interface{}
	if err := db.Raw(query).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query view: %w", err)
	}

	items := make([]RecordResult, 0, len(records))
	for _, record := range records {
		items = append(items, mapToRecordResult(record, nil))
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &ListResult{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

// validateViewQuery 验证视图查询的安全性
func validateViewQuery(query string) error {
	upperQuery := strings.ToUpper(query)

	dangerousKeywords := []string{
		"INSERT ", "UPDATE ", "DELETE ", "DROP ", "TRUNCATE ",
		"ALTER ", "CREATE ", "REPLACE ", "GRANT ", "REVOKE ",
		"EXEC ", "EXECUTE ", "CALL ", "INTO OUTFILE", "INTO DUMPFILE",
		"LOAD DATA", "LOAD XML",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperQuery, keyword) {
			return fmt.Errorf("query contains forbidden keyword: %s", strings.TrimSpace(keyword))
		}
	}

	trimmedQuery := strings.TrimSpace(upperQuery)
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return fmt.Errorf("view query must start with SELECT")
	}

	return nil
}
