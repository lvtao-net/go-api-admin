package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/validator"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 字典项缓存
var dictionaryItemsCache sync.Map

// Event 事件结构
type Event struct {
	Action     string
	Collection string
	RecordID   uint64
	Record     map[string]interface{}
}

// EventBus 事件总线
type EventBus struct {
	mu       sync.RWMutex
	handlers []chan Event
}

var eventBus = &EventBus{
	handlers: make([]chan Event, 0),
}

// Subscribe 订阅事件
func Subscribe() chan Event {
	ch := make(chan Event, 100)
	eventBus.mu.Lock()
	eventBus.handlers = append(eventBus.handlers, ch)
	eventBus.mu.Unlock()
	return ch
}

// Publish 发布事件
func Publish(event Event) {
	eventBus.mu.RLock()
	defer eventBus.mu.RUnlock()
	for _, handler := range eventBus.handlers {
		select {
		case handler <- event:
		default:
		}
	}
}

type RecordService struct {
	db *gorm.DB
}

func NewRecordService() *RecordService {
	return &RecordService{
		db: database.GetDB(),
	}
}

// RecordResult 记录结果
type RecordResult struct {
	ID      uint64                 `json:"id"`
	Created time.Time              `json:"created"`
	Updated time.Time              `json:"updated"`
	Data    map[string]interface{} `json:"-"`
}

// 时间格式常量
const DateTimeFormat = "2006-01-02 15:04:05"

// MarshalJSON 自定义JSON序列化
func (r RecordResult) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"id":      r.ID,
		"created": r.Created.Format(DateTimeFormat),
		"updated": r.Updated.Format(DateTimeFormat),
	}
	for k, v := range r.Data {
		result[k] = v
	}
	return json.Marshal(result)
}

// ListRecordsRequest 列表请求
type ListRecordsRequest struct {
	Page    int    `form:"page"`
	PerPage int    `form:"perPage"`
	Filter  string `form:"filter"`
	Sort    string `form:"sort"`
	Expand  string `form:"expand"`
}

// ListResult 列表结果
type ListResult struct {
	Page       int            `json:"page"`
	PerPage    int            `json:"perPage"`
	TotalItems int64          `json:"totalItems"`
	TotalPages int            `json:"totalPages"`
	Items      []RecordResult `json:"items"`
}

// List 获取记录列表
func (s *RecordService) List(collection *model.Collection, req *ListRecordsRequest) (*ListResult, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 500 {
		req.PerPage = 30
	}

	tableName := collection.Name
	query := s.db.Table(tableName)

	// 软删除过滤（排除已删除的记录）
	if collection.Type != "view" {
		query = query.Where("deleted_at IS NULL")
	}

	// 应用过滤条件
	if req.Filter != "" {
		conditions, args := parseFilter(req.Filter)
		if conditions != "" {
			query = query.Where(conditions, args...)
		}
	}

	// 应用排序
	if req.Sort != "" {
		order := parseSort(req.Sort)
		query = query.Order(order)
	} else if collection.Type != "view" {
		// 视图集合不默认排序，因为可能没有 created 字段
		query = query.Order("created DESC")
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// 如果没有记录，直接返回空结果
	if total == 0 {
		return &ListResult{
			Page:       req.Page,
			PerPage:    req.PerPage,
			TotalItems: 0,
			TotalPages: 0,
			Items:      []RecordResult{},
		}, nil
	}

	// 分页查询
	offset := (req.Page - 1) * req.PerPage
	var records []map[string]interface{}
	if err := query.Offset(offset).Limit(req.PerPage).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// 计算总页数
	totalPages := int(total) / req.PerPage
	if int(total)%req.PerPage > 0 {
		totalPages++
	}

	// 转换结果
	items := make([]RecordResult, 0, len(records))
	for _, record := range records {
		result := mapToRecordResult(record, collection.Fields)
		// 处理字典字段label
		s.enrichWithDictionaryLabels(collection, &result)
		// 处理关联数据展开
		if req.Expand != "" {
			if err := s.expandRecord(collection, &result, req.Expand); err != nil {
				return nil, fmt.Errorf("failed to expand record: %w", err)
			}
		}
		items = append(items, result)
	}

	return &ListResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

// expandRecord 展开关联数据
func (s *RecordService) expandRecord(collection *model.Collection, record *RecordResult, expand string) error {
	expands := strings.Split(expand, ",")

	for _, fieldName := range expands {
		fieldName = strings.TrimSpace(fieldName)
		if fieldName == "" {
			continue
		}

		// 查找关联字段
		var relationField *model.CollectionField
		for i := range collection.Fields {
			if collection.Fields[i].Name == fieldName && collection.Fields[i].Type == "relation" {
				relationField = &collection.Fields[i]
				break
			}
		}

		if relationField == nil {
			continue
		}

		// 获取关联的集合
		relationCollection := relationField.RelationCollection
		if relationCollection == "" {
			continue
		}

		// 获取关联集合信息
		relatedColl, err := NewCollectionService().GetByName(relationCollection)
		if err != nil {
			continue
		}

		// 获取关联字段值
		relationValue := record.Data[fieldName]
		if relationValue == nil {
			continue
		}

		// 处理多对多关系（JSON 数组）
		if relationField.RelationMax > 1 {
			var relatedIDs []uint64
			switch v := relationValue.(type) {
			case string:
				if v != "" {
					if err := json.Unmarshal([]byte(v), &relatedIDs); err != nil {
						continue
					}
				}
			case []interface{}:
				for _, item := range v {
					switch id := item.(type) {
					case uint64:
						relatedIDs = append(relatedIDs, id)
					case float64:
						relatedIDs = append(relatedIDs, uint64(id))
					}
				}
			}

			var expandedRecords []RecordResult
			for _, id := range relatedIDs {
				relatedRecord, err := s.GetByID(relatedColl, id)
				if err != nil {
					continue
				}
				expandedRecords = append(expandedRecords, *relatedRecord)
			}
			record.Data[fieldName] = expandedRecords
		} else {
			// 一对多或一对一关系
			var relatedID uint64
			switch v := relationValue.(type) {
			case uint64:
				relatedID = v
			case uint32:
				relatedID = uint64(v)
			case uint:
				relatedID = uint64(v)
			case float64:
				relatedID = uint64(v)
			case string:
				fmt.Sscanf(v, "%d", &relatedID)
			}
			if relatedID > 0 {
				relatedRecord, err := s.GetByID(relatedColl, relatedID)
				if err != nil {
					continue
				}
				record.Data[fieldName] = relatedRecord
			}
		}
	}

	return nil
}

// GetByID 获取单条记录
func (s *RecordService) GetByID(collection *model.Collection, id uint64) (*RecordResult, error) {
	tableName := collection.Name
	var records []map[string]interface{}

	query := s.db.Table(tableName).Where("id = ?", id)
	// 软删除过滤（排除已删除的记录）
	if collection.Type != "view" {
		query = query.Where("deleted_at IS NULL")
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("record not found")
	}

	record := records[0]
	result := mapToRecordResult(record, collection.Fields)
	// 处理字典字段label
	s.enrichWithDictionaryLabels(collection, &result)
	return &result, nil
}

// GetByField 通过指定字段获取单条记录
func (s *RecordService) GetByField(collection *model.Collection, field string, value interface{}) (*RecordResult, error) {
	tableName := collection.Name
	var records []map[string]interface{}

	query := s.db.Table(tableName).Where(fmt.Sprintf("%s = ?", field), value)
	// 软删除过滤（排除已删除的记录）
	if collection.Type != "view" {
		query = query.Where("deleted_at IS NULL")
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("record not found")
	}

	record := records[0]
	result := mapToRecordResult(record, collection.Fields)
	// 处理字典字段label
	s.enrichWithDictionaryLabels(collection, &result)
	return &result, nil
}

// Create 创建记录
func (s *RecordService) Create(collection *model.Collection, data map[string]interface{}) (*RecordResult, error) {
	tableName := collection.Name

	// 执行字段验证
	if validationErrors := s.validateRecordData(collection, data); len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %s", formatValidationErrors(validationErrors))
	}

	// 验证唯一性字段
	for _, field := range collection.Fields {
		if field.Unique {
			if value, ok := data[field.Name]; ok && value != nil && value != "" {
				var count int64
				if err := s.db.Table(tableName).Where(fmt.Sprintf("%s = ?", field.Name), value).Count(&count).Error; err != nil {
					return nil, fmt.Errorf("failed to check uniqueness: %w", err)
				}
				if count > 0 {
					return nil, fmt.Errorf("field '%s' must be unique, value '%v' already exists", field.Name, value)
				}
			}
		}
	}

	now := time.Now()

	// 构建记录数据（ID由数据库自动生成）
	record := map[string]interface{}{
		"created": now,
		"updated": now,
	}

	// 添加字段数据
	for _, field := range collection.Fields {
		if value, ok := data[field.Name]; ok {
			record[field.Name] = value
		} else if field.DefaultValue != nil {
			record[field.Name] = field.DefaultValue
		}
	}

	// 如果是Auth类型，处理密码
	if collection.Type == "auth" {
		if password, ok := data["password"].(string); ok && password != "" {
			// 密码加密
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			record["password"] = string(hashedPassword)
		}
		record["tokenKey"] = generateTokenKey()
		if _, ok := data["emailVisibility"]; !ok {
			record["emailVisibility"] = false
		}
		if _, ok := data["verified"]; !ok {
			record["verified"] = false
		}
	}

	// 插入数据库
	if err := s.db.Table(tableName).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// 获取生成的ID
	var id uint64
	if v, ok := record["id"]; ok {
		switch idVal := v.(type) {
		case uint64:
			id = idVal
		case uint32:
			id = uint64(idVal)
		case uint:
			id = uint64(idVal)
		case float64:
			id = uint64(idVal)
		}
	}

	// 获取创建的记录
	result, err := s.GetByID(collection, id)
	if err == nil {
		// 广播创建事件
		Publish(Event{
			Action:     "create",
			Collection: collection.Name,
			RecordID:   id,
			Record:     result.Data,
		})
	}

	return result, err
}

// Update 更新记录
func (s *RecordService) Update(collection *model.Collection, id uint64, data map[string]interface{}) (*RecordResult, error) {
	tableName := collection.Name

	// 检查记录是否存在（排除已删除的记录）
	var exists int64
	if err := s.db.Table(tableName).Where("id = ? AND deleted_at IS NULL", id).Count(&exists).Error; err != nil {
		return nil, fmt.Errorf("failed to check record: %w", err)
	}
	if exists == 0 {
		return nil, errors.New("record not found")
	}

	// 执行字段验证（更新时只验证提交的字段）
	if validationErrors := s.validateRecordDataForUpdate(collection, data); len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %s", formatValidationErrors(validationErrors))
	}

	// 检查集合级别的更新限制
	if len(collection.UpdateFields) > 0 {
		// 只允许更新指定的字段
		filteredData := make(map[string]interface{})
		for _, allowedField := range collection.UpdateFields {
			if value, ok := data[allowedField]; ok {
				filteredData[allowedField] = value
			}
		}
		data = filteredData
	}

	// 构建更新数据
	updateData := map[string]interface{}{
		"updated": time.Now(),
	}

	// 添加字段数据，同时检查字段级别的更新限制
	for _, field := range collection.Fields {
		if value, ok := data[field.Name]; ok {
			// 检查字段是否设置了 UpdateOnly
			if field.UpdateOnly {
				// 如果设置了 UpdateOnly，只允许更新这一个字段
				// 清空之前的数据，只保留这个字段
				updateData = map[string]interface{}{
					"updated": time.Now(),
					field.Name: value,
				}
				break
			}
			updateData[field.Name] = value
		}
	}

	// 更新数据库
	if err := s.db.Table(tableName).Where("id = ?", id).Updates(&updateData).Error; err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// 获取更新后的记录
	result, err := s.GetByID(collection, id)
	if err == nil {
		// 广播更新事件
		Publish(Event{
			Action:     "update",
			Collection: collection.Name,
			RecordID:   id,
			Record:     result.Data,
		})
	}

	return result, err
}

// Delete 删除记录（软删除）
func (s *RecordService) Delete(collection *model.Collection, id uint64) error {
	tableName := collection.Name

	// 使用软删除：设置 deleted_at 为当前时间
	result := s.db.Table(tableName).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	// 广播删除事件
	Publish(Event{
		Action:     "delete",
		Collection: collection.Name,
		RecordID:   id,
		Record:     nil,
	})

	return nil
}

// BatchDelete 批量删除（软删除）
func (s *RecordService) BatchDelete(collection *model.Collection, ids []uint64) error {
	if len(ids) == 0 {
		return errors.New("no ids provided")
	}

	tableName := collection.Name
	// 使用软删除：批量设置 deleted_at 为当前时间
	return s.db.Table(tableName).Where("id IN ? AND deleted_at IS NULL", ids).Update("deleted_at", time.Now()).Error
}

// parseFilter 安全解析过滤条件
// 支持的操作符: =, !=, >, <, >=, <=, ~ (LIKE), !~ (NOT LIKE)
// 支持的逻辑运算符: && (AND), || (OR)
// 例如: "status = 'active' && created > '2024-01-01'"
// 返回参数化的SQL条件语句和参数值，防止SQL注入
func parseFilter(filter string) (string, []interface{}) {
	if filter == "" {
		return "", nil
	}

	parser := newFilterParser(filter)
	return parser.parse()
}

// filterParser 过滤条件解析器
type filterParser struct {
	input   string
	pos     int
	tokens  []filterToken
	current int
}

// filterToken 词法单元
type filterToken struct {
	typ   tokenType
	value string
}

// tokenType 词法单元类型
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent      // 标识符（字段名）
	tokenString     // 字符串值
	tokenNumber     // 数字值
	tokenOperator   // 操作符: =, !=, >, <, >=, <=, ~, !~
	tokenAnd        // &&
	tokenOr         // ||
	tokenLParen     // (
	tokenRParen     // )
)

// newFilterParser 创建解析器
func newFilterParser(input string) *filterParser {
	p := &filterParser{input: input}
	p.tokenize()
	return p
}

// tokenize 词法分析
func (p *filterParser) tokenize() {
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		ch := p.input[p.pos]

		switch {
		case ch == '(':
			p.tokens = append(p.tokens, filterToken{tokenLParen, "("})
			p.pos++
		case ch == ')':
			p.tokens = append(p.tokens, filterToken{tokenRParen, ")"})
			p.pos++
		case ch == '&' && p.peek(1) == '&':
			p.tokens = append(p.tokens, filterToken{tokenAnd, "&&"})
			p.pos += 2
		case ch == '|' && p.peek(1) == '|':
			p.tokens = append(p.tokens, filterToken{tokenOr, "||"})
			p.pos += 2
		case ch == '!' && p.peek(1) == '=':
			p.tokens = append(p.tokens, filterToken{tokenOperator, "!="})
			p.pos += 2
		case ch == '!' && p.peek(1) == '~':
			p.tokens = append(p.tokens, filterToken{tokenOperator, "!~"})
			p.pos += 2
		case ch == '>' && p.peek(1) == '=':
			p.tokens = append(p.tokens, filterToken{tokenOperator, ">="})
			p.pos += 2
		case ch == '<' && p.peek(1) == '=':
			p.tokens = append(p.tokens, filterToken{tokenOperator, "<="})
			p.pos += 2
		case ch == '=' || ch == '>' || ch == '<' || ch == '~':
			p.tokens = append(p.tokens, filterToken{tokenOperator, string(ch)})
			p.pos++
		case ch == '\'' || ch == '"':
			p.readString(ch)
		case isDigit(ch) || (ch == '-' && p.pos+1 < len(p.input) && isDigit(p.input[p.pos+1])):
			p.readNumber()
		case isLetter(ch) || ch == '_':
			p.readIdent()
		default:
			p.pos++ // 跳过未知字符
		}
	}
	p.tokens = append(p.tokens, filterToken{tokenEOF, ""})
}

// peek 查看后续字符
func (p *filterParser) peek(offset int) byte {
	if p.pos+offset < len(p.input) {
		return p.input[p.pos+offset]
	}
	return 0
}

// skipWhitespace 跳过空白字符
func (p *filterParser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

// readString 读取字符串
func (p *filterParser) readString(quote byte) {
	p.pos++ // 跳过开始引号
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != quote {
		if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) {
			p.pos++ // 跳过转义字符
		}
		p.pos++
	}
	value := p.input[start:p.pos]
	if p.pos < len(p.input) {
		p.pos++ // 跳过结束引号
	}
	p.tokens = append(p.tokens, filterToken{tokenString, value})
}

// readNumber 读取数字
func (p *filterParser) readNumber() {
	start := p.pos
	if p.pos < len(p.input) && p.input[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.input) && (isDigit(p.input[p.pos]) || p.input[p.pos] == '.') {
		p.pos++
	}
	p.tokens = append(p.tokens, filterToken{tokenNumber, p.input[start:p.pos]})
}

// readIdent 读取标识符
func (p *filterParser) readIdent() {
	start := p.pos
	for p.pos < len(p.input) && (isLetter(p.input[p.pos]) || isDigit(p.input[p.pos]) || p.input[p.pos] == '_') {
		p.pos++
	}
	p.tokens = append(p.tokens, filterToken{tokenIdent, p.input[start:p.pos]})
}

// parse 解析生成SQL
func (p *filterParser) parse() (string, []interface{}) {
	if len(p.tokens) == 0 || (len(p.tokens) == 1 && p.tokens[0].typ == tokenEOF) {
		return "", nil
	}

	var args []interface{}
	cond := p.parseOr(&args)
	return cond, args
}

// parseOr 解析 OR 表达式
func (p *filterParser) parseOr(args *[]interface{}) string {
	left := p.parseAnd(args)
	for p.currentToken().typ == tokenOr {
		p.current++
		right := p.parseAnd(args)
		left = fmt.Sprintf("(%s OR %s)", left, right)
	}
	return left
}

// parseAnd 解析 AND 表达式
func (p *filterParser) parseAnd(args *[]interface{}) string {
	left := p.parsePrimary(args)
	for p.currentToken().typ == tokenAnd {
		p.current++
		right := p.parsePrimary(args)
		left = fmt.Sprintf("(%s AND %s)", left, right)
	}
	return left
}

// parsePrimary 解析基本表达式
func (p *filterParser) parsePrimary(args *[]interface{}) string {
	tok := p.currentToken()

	// 处理括号
	if tok.typ == tokenLParen {
		p.current++
		cond := p.parseOr(args)
		if p.currentToken().typ == tokenRParen {
			p.current++
		}
		return cond
	}

	// 字段名
	if tok.typ != tokenIdent {
		return "1=1" // 无效表达式，返回真
	}
	field := tok.value
	p.current++

	// 操作符
	opTok := p.currentToken()
	if opTok.typ != tokenOperator {
		return "1=1"
	}
	op := opTok.value
	p.current++

	// 值
	valTok := p.currentToken()
	p.current++

	var placeholder string
	switch valTok.typ {
	case tokenString:
		*args = append(*args, valTok.value)
		placeholder = "?"
	case tokenNumber:
		*args = append(*args, valTok.value)
		placeholder = "?"
	default:
		return "1=1"
	}

	// 转换操作符
	sqlOp := p.convertOperator(op)

	// 验证字段名（只允许字母、数字、下划线）
	if !isValidFieldName(field) {
		return "1=1"
	}

	return fmt.Sprintf("%s %s %s", field, sqlOp, placeholder)
}

// currentToken 获取当前token
func (p *filterParser) currentToken() filterToken {
	if p.current < len(p.tokens) {
		return p.tokens[p.current]
	}
	return filterToken{tokenEOF, ""}
}

// convertOperator 转换操作符为SQL
func (p *filterParser) convertOperator(op string) string {
	switch op {
	case "~":
		return "LIKE"
	case "!~":
		return "NOT LIKE"
	default:
		return op
	}
}

// isValidFieldName 验证字段名是否合法
func isValidFieldName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if i == 0 {
			if !isLetter(ch) && ch != '_' {
				return false
			}
		} else {
			if !isLetter(ch) && !isDigit(ch) && ch != '_' {
				return false
			}
		}
	}
	return true
}

// isLetter 判断是否为字母
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit 判断是否为数字
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isWhitespace 判断是否为空白字符
func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// parseSort 解析排序条件
// 例如: "-created,name" 表示 created DESC, name ASC
func parseSort(sort string) string {
	if sort == "" {
		return "created DESC"
	}

	parts := strings.Split(sort, ",")
	var orders []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, "-") {
			orders = append(orders, fmt.Sprintf("%s DESC", strings.TrimPrefix(part, "-")))
		} else {
			orders = append(orders, fmt.Sprintf("%s ASC", part))
		}
	}

	return strings.Join(orders, ", ")
}

// mapToRecordResult 将map转换为RecordResult
func mapToRecordResult(m map[string]interface{}, fields []model.CollectionField) RecordResult {
	result := RecordResult{
		Data: make(map[string]interface{}),
	}

	// 构建字段格式化映射
	fieldFormats := make(map[string]string)
	for _, field := range fields {
		if field.DateFormat != "" {
			fieldFormats[field.Name] = field.DateFormat
		}
	}

	for k, v := range m {
		switch k {
		case "id":
			switch id := v.(type) {
			case uint64:
				result.ID = id
			case uint32:
				result.ID = uint64(id)
			case uint:
				result.ID = uint64(id)
			case float64:
				result.ID = uint64(id)
			case string:
				// 兼容旧的string类型，尝试转换
				var parsed uint64
				fmt.Sscanf(id, "%d", &parsed)
				result.ID = parsed
			}
		case "created":
			if t, ok := v.(time.Time); ok {
				result.Created = t
			}
		case "updated":
			if t, ok := v.(time.Time); ok {
				result.Updated = t
			}
		default:
			// 处理日期字段的格式化
			if format, ok := fieldFormats[k]; ok {
				result.Data[k] = formatDateTimeValue(v, format)
			} else {
				result.Data[k] = v
			}
		}
	}

	return result
}

// formatDateTimeValue 根据格式化选项格式化日期时间值
func formatDateTimeValue(v interface{}, format string) interface{} {
	if v == nil {
		return nil
	}

	var t time.Time
	switch val := v.(type) {
	case time.Time:
		t = val
	case string:
		// 尝试解析字符串时间
		parsed, err := time.Parse(time.RFC3339, val)
		if err != nil {
			parsed, err = time.Parse(DateTimeFormat, val)
			if err != nil {
				return v // 无法解析，返回原值
			}
		}
		t = parsed
	default:
		return v
	}

	// 根据格式化选项输出
	switch format {
	case "date":
		return t.Format("2006-01-02")
	case "datetime":
		return t.Format(DateTimeFormat)
	case "time":
		return t.Format("15:04:05")
	default:
		// 自定义格式
		return t.Format(format)
	}
}

// generateTokenKey 生成TokenKey
func generateTokenKey() string {
	return generateID() + generateID()
}

// filterFieldsByPermission 根据字段权限过滤数据
func filterFieldsByPermission(data map[string]interface{}, fields []model.CollectionField, isAdmin bool) map[string]interface{} {
	if isAdmin {
		return data
	}

	filtered := make(map[string]interface{})
	for k, v := range data {
		// 始终保留系统字段
		if k == "id" || k == "created" || k == "updated" {
			filtered[k] = v
			continue
		}

		// 查找字段定义
		var field model.CollectionField
		found := false
		for _, f := range fields {
			if f.Name == k {
				field = f
				found = true
				break
			}
		}

		if !found {
			// 未定义的字段，默认可见
			filtered[k] = v
			continue
		}

		// 检查 API 级别权限
		if field.APIDisabled {
			// API 禁止访问，完全隐藏
			continue
		}

		if field.APIWriteOnly {
			// 只写字段，在读取时隐藏（如密码字段）
			continue
		}

		// 检查是否隐藏
		if field.Hidden {
			continue
		}

		filtered[k] = v
	}

	return filtered
}

// FilterFieldsForList 根据列表 API 权限过滤数据
func FilterFieldsForList(data map[string]interface{}, fields []model.CollectionField, isAdmin bool) map[string]interface{} {
	if isAdmin {
		return data
	}

	filtered := make(map[string]interface{})
	for k, v := range data {
		// 始终保留系统字段
		if k == "id" || k == "created" || k == "updated" {
			filtered[k] = v
			continue
		}

		// 查找字段定义
		var field model.CollectionField
		found := false
		for _, f := range fields {
			if f.Name == k {
				field = f
				found = true
				break
			}
		}

		if !found {
			filtered[k] = v
			continue
		}

		// 检查 API 级别权限
		if field.APIDisabled || field.APIWriteOnly || field.APIHiddenList {
			continue
		}

		if field.Hidden {
			continue
		}

		filtered[k] = v
	}

	return filtered
}

// FilterFieldsForView 根据详情 API 权限过滤数据
func FilterFieldsForView(data map[string]interface{}, fields []model.CollectionField, isAdmin bool) map[string]interface{} {
	if isAdmin {
		return data
	}

	filtered := make(map[string]interface{})
	for k, v := range data {
		// 始终保留系统字段
		if k == "id" || k == "created" || k == "updated" {
			filtered[k] = v
			continue
		}

		// 查找字段定义
		var field model.CollectionField
		found := false
		for _, f := range fields {
			if f.Name == k {
				field = f
				found = true
				break
			}
		}

		if !found {
			filtered[k] = v
			continue
		}

		// 检查 API 级别权限
		if field.APIDisabled || field.APIWriteOnly || field.APIHiddenView {
			continue
		}

		if field.Hidden {
			continue
		}

		filtered[k] = v
	}

	return filtered
}

// FilterFieldsForCreate 根据创建权限过滤数据
func FilterFieldsForCreate(data map[string]interface{}, fields []model.CollectionField, isAdmin bool) map[string]interface{} {
	if isAdmin {
		return data
	}

	filtered := make(map[string]interface{})

	for _, field := range fields {
		// 跳过 API 禁止的字段
		if field.APIDisabled {
			continue
		}

		// 跳过只读字段
		if field.APIReadOnly {
			continue
		}

		// 跳过隐藏字段
		if field.Hidden || field.HiddenOnForm {
			continue
		}

		// 注意：Editable 默认值是 false，但我们应该允许创建必填字段
		// 只有显式设置 Editable = false 且不是必填字段时才跳过
		// 系统字段（id, created, updated）不需要用户输入

		// 如果数据中存在该字段，添加到过滤后的数据中
		if value, ok := data[field.Name]; ok {
			filtered[field.Name] = value
		} else if field.DefaultValue != nil {
			filtered[field.Name] = field.DefaultValue
		}
	}

	return filtered
}

// FilterFieldsForUpdate 根据更新权限过滤数据
func FilterFieldsForUpdate(data map[string]interface{}, fields []model.CollectionField, isAdmin bool) map[string]interface{} {
	if isAdmin {
		return data
	}

	filtered := make(map[string]interface{})

	for _, field := range fields {
		// 跳过 API 禁止的字段
		if field.APIDisabled {
			continue
		}

		// 跳过只读字段
		if field.APIReadOnly {
			continue
		}

		// 跳过隐藏字段
		if field.Hidden {
			continue
		}

		// 如果数据中存在该字段，添加到过滤后的数据中
		if value, ok := data[field.Name]; ok {
			filtered[field.Name] = value
		}
	}

	return filtered
}

// getDictionaryLabel 获取字典项的label
func getDictionaryLabel(dictionaryName, value string) (string, error) {
	// 尝试从缓存获取
	cacheKey := dictionaryName
	if cached, ok := dictionaryItemsCache.Load(cacheKey); ok {
		if items, ok := cached.(map[string]string); ok {
			if label, exists := items[value]; exists {
				return label, nil
			}
		}
	}

	// 从数据库查询字典项
	var items []model.DictionaryItem
	db := database.GetDB()
	if err := db.Where("dictionary_name = ?", dictionaryName).Find(&items).Error; err != nil {
		return "", err
	}

	// 构建缓存
	itemMap := make(map[string]string)
	for _, item := range items {
		itemMap[item.Value] = item.Label
	}
	dictionaryItemsCache.Store(cacheKey, itemMap)

	if label, exists := itemMap[value]; exists {
		return label, nil
	}
	return value, nil
}

// enrichWithDictionaryLabels 为记录添加字典label
func (s *RecordService) enrichWithDictionaryLabels(collection *model.Collection, record *RecordResult) {
	for _, field := range collection.Fields {
		// 只处理单选、多选、下拉字段，且关联了字典的
		if !isDictionaryFieldType(field.Type) || field.Dictionary == "" {
			continue
		}

		value := record.Data[field.Name]
		if value == nil {
			continue
		}

		switch field.Type {
		case "select", "radio":
			// 单选：value是字符串
			if strVal, ok := value.(string); ok && strVal != "" {
				label, _ := getDictionaryLabel(field.Dictionary, strVal)
				record.Data[field.Name+"_label"] = label
			}
		case "checkbox":
			// 多选：value可能是JSON数组字符串
			var values []string
			switch v := value.(type) {
			case string:
				if v != "" {
					if err := json.Unmarshal([]byte(v), &values); err != nil {
						// 如果解析失败，尝试作为单个值处理
						values = []string{v}
					}
				}
			case []interface{}:
				for _, item := range v {
					if strItem, ok := item.(string); ok {
						values = append(values, strItem)
					}
				}
			case []string:
				values = v
			}

			if len(values) > 0 {
				labels := make([]string, 0, len(values))
				for _, val := range values {
					label, _ := getDictionaryLabel(field.Dictionary, val)
					labels = append(labels, label)
				}
				record.Data[field.Name+"_label"] = strings.Join(labels, ", ")
			}
		}
	}
}

// isDictionaryFieldType 判断字段类型是否为字典类型
func isDictionaryFieldType(fieldType string) bool {
	return fieldType == "select" || fieldType == "radio" || fieldType == "checkbox"
}

// validateRecordData 验证记录数据
func (s *RecordService) validateRecordData(collection *model.Collection, data map[string]interface{}) []validator.ValidationError {
	// 将 CollectionField 转换为 FieldValidator 接口
	fieldValidators := make([]validator.FieldValidator, 0, len(collection.Fields))
	for _, field := range collection.Fields {
		fieldValidators = append(fieldValidators, validator.CollectionFieldAdapter{
			Name:               field.Name,
			Required:           field.Required,
			ValidationRules:    field.ValidationRules,
			ValidationMessages: field.ValidationMessages,
			Min:                field.Min,
			Max:                field.Max,
		})
	}

	return validator.ValidateRecord(data, fieldValidators)
}

// validateRecordDataForUpdate 验证更新数据（只验证提交的字段）
func (s *RecordService) validateRecordDataForUpdate(collection *model.Collection, data map[string]interface{}) []validator.ValidationError {
	// 将 CollectionField 转换为 FieldValidator 接口
	fieldValidators := make([]validator.FieldValidator, 0, len(collection.Fields))
	for _, field := range collection.Fields {
		fieldValidators = append(fieldValidators, validator.CollectionFieldAdapter{
			Name:               field.Name,
			Required:           field.Required,
			ValidationRules:    field.ValidationRules,
			ValidationMessages: field.ValidationMessages,
			Min:                field.Min,
			Max:                field.Max,
		})
	}

	return validator.ValidateRecordForUpdate(data, fieldValidators)
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(errors []validator.ValidationError) string {
	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// ViewQuery 执行视图查询
func (s *RecordService) ViewQuery(collection *model.Collection, req *ListRecordsRequest) (*ListResult, error) {
	viewQuery := collection.ViewQuery
	if viewQuery == "" {
		return nil, errors.New("view query is empty")
	}

	// 构建分页查询
	page := req.Page
	perPage := req.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 500 {
		perPage = 30
	}
	offset := (page - 1) * perPage

	// 查询数据
	query := fmt.Sprintf("SELECT * FROM (%s) AS view_table LIMIT %d OFFSET %d", viewQuery, perPage, offset)
	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute view query: %w", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// 扫描数据
	var items []RecordResult
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		data := make(map[string]interface{})
		for i, col := range columns {
			data[col] = values[i]
		}
		items = append(items, RecordResult{Data: data})
	}

	// 查询总数
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS view_table", viewQuery)
	if err := s.db.Raw(countQuery).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count: %w", err)
	}

	return &ListResult{
		Items:       items,
		Page:        page,
		PerPage:     perPage,
		TotalItems:  total,
		TotalPages:  int((total + int64(perPage) - 1) / int64(perPage)),
	}, nil
}
