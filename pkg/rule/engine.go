package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"gorm.io/gorm"
)

// RuleEngine 规则引擎
type RuleEngine struct {
	db *gorm.DB
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// NewRuleEngineWithDB 创建带数据库连接的规则引擎（用于获取关联数据）
func NewRuleEngineWithDB(db *gorm.DB) *RuleEngine {
	return &RuleEngine{db: db}
}

// Context 规则执行上下文
type Context struct {
	AuthID         uint64                 // 当前用户ID
	AuthEmail      string                 // 当前用户邮箱
	AuthCollection string                 // 当前用户所属集合
	AuthRecord     map[string]interface{} // 当前用户记录
	Record         map[string]interface{} // 当前操作的记录
	Body           map[string]interface{} // 请求体数据
	RelatedRecords map[string]interface{} // 关联记录数据缓存
}

// SetDB 设置数据库连接（用于获取关联数据）
func (e *RuleEngine) SetDB(db *gorm.DB) {
	e.db = db
}

// Check 检查规则是否通过
func (e *RuleEngine) Check(rule string, ctx *Context) (bool, error) {
	// 处理快捷规则类型
	switch rule {
	case "public":
		// 公开访问，任何人都可以访问
		return true, nil
	case "auth":
		// 需要认证，登录用户可访问
		return ctx.AuthID > 0, nil
	case "owner":
		// 仅所有者，需要登录且 @request.auth.id = record.id
		if ctx.AuthID == 0 {
			return false, nil
		}
		if ctx.Record == nil {
			return false, nil
		}
		// 检查记录的 id 或 user_id 字段是否等于当前用户ID
		var recordID uint64
		switch v := ctx.Record["id"].(type) {
		case uint64:
			recordID = v
		case float64:
			recordID = uint64(v)
		}
		if recordID == ctx.AuthID {
			return true, nil
		}
		var userID uint64
		switch v := ctx.Record["user_id"].(type) {
		case uint64:
			userID = v
		case float64:
			userID = uint64(v)
		}
		if userID == ctx.AuthID {
			return true, nil
		}
		// 检查记录是否有 author 字段
		var authorID uint64
		switch v := ctx.Record["author"].(type) {
		case uint64:
			authorID = v
		case float64:
			authorID = uint64(v)
		}
		if authorID == ctx.AuthID {
			return true, nil
		}
		// 嵌套访问 author.id
		if author, ok := ctx.Record["author"].(map[string]interface{}); ok {
			switch v := author["id"].(type) {
			case uint64:
				if v == ctx.AuthID {
					return true, nil
				}
			case float64:
				if uint64(v) == ctx.AuthID {
					return true, nil
				}
			}
		}
		return false, nil
	case "admin":
		// 仅管理员（需要通过上下文传递）
		if isAdmin, ok := ctx.AuthRecord["is_admin"].(bool); ok && isAdmin {
			return true, nil
		}
		// 或者检查是否有管理员标记
		if ctx.AuthCollection == "_admins" {
			return true, nil
		}
		return false, nil
	case "disabled":
		// 禁用该操作
		return false, nil
	case "", "null":
		// 空规则表示允许访问（向后兼容）
		return true, nil
	}

	// 解析自定义规则表达式
	expr, err := parseRule(rule)
	if err != nil {
		return false, err
	}

	return e.evaluate(expr, ctx)
}

// evaluate 评估表达式
func (e *RuleEngine) evaluate(expr *Expression, ctx *Context) (bool, error) {
	switch expr.Type {
	case "literal":
		return expr.Value == "true", nil

	case "variable":
		return e.getVariableValue(expr.Value, ctx)

	case "comparison":
		return e.evaluateComparison(expr, ctx)

	case "logical":
		return e.evaluateLogical(expr, ctx)

	default:
		return false, fmt.Errorf("unknown expression type: %s", expr.Type)
	}
}

// getVariableValue 获取变量值
func (e *RuleEngine) getVariableValue(name string, ctx *Context) (bool, error) {
	switch name {
	case "@request.auth.id":
		return ctx.AuthID > 0, nil
	case "@request.auth.email":
		return ctx.AuthEmail != "", nil
	case "@request.auth":
		return ctx.AuthID > 0, nil
	default:
		// 支持嵌套访问：@request.record.field.subfield 或 @request.body.field.subfield
		// 例如：@request.record.author.id, @request.record.author.email

		// 处理 @request.record.xxx 格式
		if strings.HasPrefix(name, "@request.record.") {
			fieldPath := strings.TrimPrefix(name, "@request.record.")
			return e.getNestedValue(ctx.Record, fieldPath)
		}

		// 处理 @request.body.xxx 格式
		if strings.HasPrefix(name, "@request.body.") {
			fieldPath := strings.TrimPrefix(name, "@request.body.")
			return e.getNestedValue(ctx.Body, fieldPath)
		}

		// 处理简化的关联字段格式（兼容旧版本）
		// 例如：author.id 表示从当前记录的 author 字段获取关联记录的 id
		if ctx.Record != nil {
			if val, ok := ctx.Record[name]; ok {
				return val != nil && val != "", nil
			}
		}

		// 尝试从AuthRecord中获取
		if ctx.AuthRecord != nil {
			if val, ok := ctx.AuthRecord[name]; ok {
				return val != nil && val != "", nil
			}
		}

		return false, nil
	}
}

// getNestedValue 获取嵌套字段的值
// 例如：fieldPath = "author.id" 表示获取 record["author"]["id"]
func (e *RuleEngine) getNestedValue(data map[string]interface{}, fieldPath string) (bool, error) {
	if data == nil {
		return false, nil
	}

	// 支持点号分隔的路径，如 "author.id"
	parts := strings.Split(fieldPath, ".")
	if len(parts) < 1 {
		return false, nil
	}

	current := data
	for i, part := range parts {
		if current == nil {
			return false, nil
		}

		val, ok := current[part]
		if !ok {
			// 如果是最后一部分且没找到，返回false
			if i == len(parts)-1 {
				return false, nil
			}
			return false, nil
		}

		// 如果是最后一部分，返回值
		if i == len(parts) - 1 {
			return val != nil && val != "", nil
		}

		// 继续遍历下一层
		if nextMap, ok := val.(map[string]interface{}); ok {
			current = nextMap
		} else if strVal, ok := val.(string); ok && strVal != "" && e.db != nil {
			// 如果是关联字段ID且有数据库连接，尝试获取关联记录
			// 需要集合信息来确定关联哪个表，这里简化处理
			// 实际使用中应该由调用方预先加载关联数据
			current = nil
		} else {
			current = nil
		}
	}

	return false, nil
}

// GetNestedValue 获取嵌套字段的实际值（不限于布尔）
func (e *RuleEngine) GetNestedValue(data map[string]interface{}, fieldPath string) (interface{}, error) {
	if data == nil {
		return nil, nil
	}

	parts := strings.Split(fieldPath, ".")
	if len(parts) < 1 {
		return nil, nil
	}

	current := data
	for i, part := range parts {
		if current == nil {
			return nil, nil
		}

		val, ok := current[part]
		if !ok {
			return nil, nil
		}

		// 如果是最后一部分，返回值
		if i == len(parts) - 1 {
			return val, nil
		}

		// 继续遍历下一层
		if nextMap, ok := val.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil, nil
		}
	}

	return nil, nil
}

// ResolveValue 解析变量值（返回实际值而非布尔）
func (e *RuleEngine) ResolveValue(name string, ctx *Context) (interface{}, error) {
	// 处理 @request.record.xxx 格式
	if strings.HasPrefix(name, "@request.record.") {
		fieldPath := strings.TrimPrefix(name, "@request.record.")
		return e.GetNestedValue(ctx.Record, fieldPath)
	}

	// 处理 @request.body.xxx 格式
	if strings.HasPrefix(name, "@request.body.") {
		fieldPath := strings.TrimPrefix(name, "@request.body.")
		return e.GetNestedValue(ctx.Body, fieldPath)
	}

	// 处理 @request.auth.id
	if name == "@request.auth.id" {
		return ctx.AuthID, nil
	}

	// 处理 @request.auth.email
	if name == "@request.auth.email" {
		return ctx.AuthEmail, nil
	}

	// 从Record中获取
	if ctx.Record != nil {
		if val, ok := ctx.Record[name]; ok {
			return val, nil
		}
	}

	// 从AuthRecord中获取
	if ctx.AuthRecord != nil {
		if val, ok := ctx.AuthRecord[name]; ok {
			return val, nil
		}
	}

	return nil, nil
}

// LoadRelatedRecord 加载关联记录数据
// collectionName: 关联集合名称
// recordID: 关联记录ID
// 返回关联记录的map
func (e *RuleEngine) LoadRelatedRecord(collectionName string, recordID uint64) (map[string]interface{}, error) {
	if e.db == nil || collectionName == "" || recordID == 0 {
		return nil, nil
	}

	var record map[string]interface{}
	if err := e.db.Table(collectionName).Where("id = ?", recordID).Scan(&record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

// ExpandRecordRelations 展开记录中的关联字段
// 这个方法用于在获取记录后，预先加载关联数据到Record中
// 以便规则引擎可以访问 @request.record.author.id 这样的嵌套字段
func (e *RuleEngine) ExpandRecordRelations(record map[string]interface{}, collectionName string, expandFields []string) error {
	if e.db == nil || record == nil {
		return nil
	}

	for _, fieldName := range expandFields {
		if val, ok := record[fieldName]; ok {
			var relationID uint64
			switch v := val.(type) {
			case uint64:
				relationID = v
			case float64:
				relationID = uint64(v)
			}
			if relationID > 0 {
				// 查找集合中的关联字段定义
				relatedRecord, err := e.LoadRelatedRecord(collectionName, relationID)
				if err == nil && relatedRecord != nil {
					record[fieldName] = relatedRecord
				}
			}
		}
	}

	return nil
}

// evaluateComparison 评估比较表达式
func (e *RuleEngine) evaluateComparison(expr *Expression, ctx *Context) (bool, error) {
	left, err := e.resolveValue(expr.Left, ctx)
	if err != nil {
		return false, err
	}

	right, err := e.resolveValue(expr.Right, ctx)
	if err != nil {
		return false, err
	}

	switch expr.Operator {
	case "=":
		return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right), nil
	case "!=":
		return fmt.Sprintf("%v", left) != fmt.Sprintf("%v", right), nil
	case ">":
		return compareNumbers(left, right) > 0, nil
	case "<":
		return compareNumbers(left, right) < 0, nil
	case ">=":
		return compareNumbers(left, right) >= 0, nil
	case "<=":
		return compareNumbers(left, right) <= 0, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", expr.Operator)
	}
}

// evaluateLogical 评估逻辑表达式
func (e *RuleEngine) evaluateLogical(expr *Expression, ctx *Context) (bool, error) {
	switch expr.LogicalOp {
	case "&&":
		for _, child := range expr.Children {
			result, err := e.evaluate(child, ctx)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil

	case "||":
		for _, child := range expr.Children {
			result, err := e.evaluate(child, ctx)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil

	case "!":
		if len(expr.Children) > 0 {
			result, err := e.evaluate(expr.Children[0], ctx)
			return !result, err
		}
		return true, nil

	default:
		return false, fmt.Errorf("unknown logical operator: %s", expr.LogicalOp)
	}
}

// resolveValue 解析值
func (e *RuleEngine) resolveValue(val interface{}, ctx *Context) (interface{}, error) {
	str, ok := val.(string)
	if !ok {
		return val, nil
	}

	// 变量 - 使用 ResolveValue 获取实际值（支持关联字段）
	if strings.HasPrefix(str, "@") {
		return e.ResolveValue(str, ctx)
	}

	// 字符串字面量
	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
		return strings.Trim(str, "'"), nil
	}

	// 数字
	if regexp.MustCompile(`^-?\d+(\.\d+)?$`).MatchString(str) {
		return str, nil
	}

	return str, nil
}

// compareNumbers 比较数字
func compareNumbers(a, b interface{}) int {
	af, _ := toFloat64(a)
	bf, _ := toFloat64(b)
	if af > bf {
		return 1
	} else if af < bf {
		return -1
	}
	return 0
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f, nil
	default:
		return 0, errors.New("cannot convert to float64")
	}
}

// Expression 表达式
type Expression struct {
	Type      string        // literal, variable, comparison, logical
	Value     string        // 用于 literal 和 variable
	Operator  string        // 用于 comparison
	Left      interface{}  // 用于 comparison
	Right     interface{}  // 用于 comparison
	LogicalOp string       // 用于 logical: &&, ||, !
	Children  []*Expression // 用于 logical
}

// parseRule 解析规则字符串
func parseRule(rule string) (*Expression, error) {
	// 简化实现：处理常见的规则模式
	rule = strings.TrimSpace(rule)

	// 处理 && 和 ||
	if strings.Contains(rule, "&&") {
		parts := strings.Split(rule, "&&")
		children := make([]*Expression, 0, len(parts))
		for _, part := range parts {
			child, err := parseRule(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
		return &Expression{
			Type:      "logical",
			LogicalOp: "&&",
			Children:  children,
		}, nil
	}

	if strings.Contains(rule, "||") {
		parts := strings.Split(rule, "||")
		children := make([]*Expression, 0, len(parts))
		for _, part := range parts {
			child, err := parseRule(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
		return &Expression{
			Type:      "logical",
			LogicalOp: "||",
			Children:  children,
		}, nil
	}

	// 处理 !
	if strings.HasPrefix(rule, "!") {
		child, err := parseRule(strings.TrimSpace(rule[1:]))
		if err != nil {
			return nil, err
		}
		return &Expression{
			Type:      "logical",
			LogicalOp: "!",
			Children:  []*Expression{child},
		}, nil
	}

	// 处理比较运算
	for _, op := range []string{"!=", ">=", "<=", "=", ">", "<"} {
		idx := strings.Index(rule, op)
		if idx > 0 {
			left := strings.TrimSpace(rule[:idx])
			right := strings.TrimSpace(rule[idx+len(op):])
			return &Expression{
				Type:     "comparison",
				Operator: op,
				Left:     left,
				Right:    right,
			}, nil
		}
	}

	// 处理变量或字面量
	if strings.HasPrefix(rule, "@") {
		return &Expression{
			Type:  "variable",
			Value: rule,
		}, nil
	}

	// 处理布尔字面量
	if rule == "true" || rule == "false" {
		return &Expression{
			Type:  "literal",
			Value: rule,
		}, nil
	}

	// 默认为变量
	return &Expression{
		Type:  "variable",
		Value: rule,
	}, nil
}

// CheckAuthRule 检查认证规则
func CheckAuthRule(rule string, token string) (*Context, error) {
	ctx := &Context{}

	if token == "" {
		// 未登录，只能访问公共资源
		return ctx, nil
	}

	// 解析Token
	claims, err := auth.ValidateUserToken(token)
	if err != nil {
		return ctx, nil // Token无效，视为未登录
	}

	ctx.AuthEmail = claims.Email
	ctx.AuthCollection = claims.Collection

	return ctx, nil
}
