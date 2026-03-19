package transaction

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
	"github.com/lvtao/go-gin-api-admin/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Engine 事务执行引擎
type Engine struct {
	db *gorm.DB
}

// NewEngine 创建事务执行引擎
func NewEngine() *Engine {
	return &Engine{
		db: database.GetDB(),
	}
}

// Execute 执行事务
func (e *Engine) Execute(collection *model.Collection, params map[string]interface{}, userID uint) (*TransactionResult, error) {
	if collection.Type != "transaction" {
		return nil, fmt.Errorf("集合类型不是事务类型")
	}

	if len(collection.TransactionSteps) == 0 {
		return nil, fmt.Errorf("事务步骤为空")
	}

	ctx := &Context{
		Params:      params,
		UserID:      userID,
		Results:     make(map[string]interface{}),
		InsertedIDs: make(map[string]uint),
	}

	var result *TransactionResult
	var err error

	// 使用数据库事务
	txErr := e.db.Transaction(func(tx *gorm.DB) error {
		result, err = e.executeSteps(tx, collection.TransactionSteps, ctx)
		return err
	})

	if txErr != nil {
		logger.Error("Transaction failed",
			zap.String("collection", collection.Name),
			zap.Uint("userID", userID),
			zap.Any("params", params),
			zap.Error(txErr),
		)
		return nil, txErr
	}

	logger.Info("Transaction completed",
		zap.String("collection", collection.Name),
		zap.Uint("userID", userID),
	)

	return result, nil
}

// executeSteps 执行步骤列表
func (e *Engine) executeSteps(tx *gorm.DB, steps []model.TransactionStep, ctx *Context) (*TransactionResult, error) {
	for i, step := range steps {
		logger.Debug("Executing step",
			zap.Int("step", i+1),
			zap.String("type", step.Type),
			zap.String("table", step.Table),
		)

		switch step.Type {
		case "query":
			if err := e.executeQuery(tx, &step, ctx); err != nil {
				return nil, err
			}

		case "validate":
			if err := e.executeValidate(&step, ctx); err != nil {
				return nil, err
			}

		case "update":
			if err := e.executeUpdate(tx, &step, ctx); err != nil {
				return nil, err
			}

		case "insert":
			if err := e.executeInsert(tx, &step, ctx); err != nil {
				return nil, err
			}

		case "delete":
			if err := e.executeDelete(tx, &step, ctx); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("未知步骤类型: %s", step.Type)
		}
	}

	return &TransactionResult{
		Success: true,
		Message: "执行成功",
		Data:    ctx.Results,
	}, nil
}

// executeQuery 执行查询步骤
func (e *Engine) executeQuery(tx *gorm.DB, step *model.TransactionStep, ctx *Context) error {
	if step.Table == "" {
		return fmt.Errorf("查询步骤需要指定表名")
	}

	query := tx.Table(step.Table)

	// 构建查询条件
	for _, cond := range step.Conditions {
		value := e.resolveValue(cond.Value, cond.ValueFrom, ctx)
		query = e.applyCondition(query, cond.Field, cond.Operator, value)
	}

	// 执行查询
	var result map[string]interface{}
	err := query.First(&result).Error

	if err == gorm.ErrRecordNotFound {
		if step.Required {
			if step.Error != "" {
				return fmt.Errorf("%s", step.Error)
			}
			return fmt.Errorf("记录不存在: %s", step.Table)
		}
		// 非必须，设置为 nil
		if step.Alias != "" {
			ctx.Results[step.Alias] = nil
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}

	// 保存结果
	if step.Alias != "" {
		ctx.Results[step.Alias] = result
	}

	return nil
}

// executeValidate 执行验证步骤
func (e *Engine) executeValidate(step *model.TransactionStep, ctx *Context) error {
	if step.ValidateCondition == "" {
		return nil
	}

	// 解析验证条件
	result := e.evaluateCondition(step.ValidateCondition, ctx)

	if !result {
		if step.Error != "" {
			return fmt.Errorf("%s", step.Error)
		}
		return fmt.Errorf("验证失败: %s", step.ValidateCondition)
	}

	return nil
}

// executeUpdate 执行更新步骤
func (e *Engine) executeUpdate(tx *gorm.DB, step *model.TransactionStep, ctx *Context) error {
	if step.Table == "" {
		return fmt.Errorf("更新步骤需要指定表名")
	}

	// 解析更新数据
	data := make(map[string]interface{})
	for k, v := range step.Data {
		data[k] = e.resolveValue(v, "", ctx)
	}
	data["updated"] = time.Now()

	query := tx.Table(step.Table)

	// 构建条件
	for _, cond := range step.Conditions {
		value := e.resolveValue(cond.Value, cond.ValueFrom, ctx)
		query = e.applyCondition(query, cond.Field, cond.Operator, value)
	}

	result := query.Updates(data)
	if result.Error != nil {
		return fmt.Errorf("更新失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		if step.Error != "" {
			return fmt.Errorf("%s", step.Error)
		}
		return fmt.Errorf("未更新任何记录")
	}

	return nil
}

// executeInsert 执行插入步骤
func (e *Engine) executeInsert(tx *gorm.DB, step *model.TransactionStep, ctx *Context) error {
	if step.Table == "" {
		return fmt.Errorf("插入步骤需要指定表名")
	}

	// 解析插入数据
	data := make(map[string]interface{})
	for k, v := range step.Data {
		data[k] = e.resolveValue(v, "", ctx)
	}
	data["created"] = time.Now()
	data["updated"] = time.Now()

	result := tx.Table(step.Table).Create(&data)
	if result.Error != nil {
		return fmt.Errorf("插入失败: %w", result.Error)
	}

	// 获取插入的ID
	if id, ok := data["id"]; ok {
		switch v := id.(type) {
		case uint:
			if step.Alias != "" {
				ctx.InsertedIDs[step.Alias] = v
			}
		case float64:
			if step.Alias != "" {
				ctx.InsertedIDs[step.Alias] = uint(v)
			}
		case int:
			if step.Alias != "" {
				ctx.InsertedIDs[step.Alias] = uint(v)
			}
		}
	}

	// 保存结果
	if step.Alias != "" {
		ctx.Results[step.Alias] = data
	}

	return nil
}

// executeDelete 执行删除步骤
func (e *Engine) executeDelete(tx *gorm.DB, step *model.TransactionStep, ctx *Context) error {
	if step.Table == "" {
		return fmt.Errorf("删除步骤需要指定表名")
	}

	query := tx.Table(step.Table)

	// 构建条件
	for _, cond := range step.Conditions {
		value := e.resolveValue(cond.Value, cond.ValueFrom, ctx)
		query = e.applyCondition(query, cond.Field, cond.Operator, value)
	}

	result := query.Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除失败: %w", result.Error)
	}

	return nil
}

// applyCondition 应用查询条件
func (e *Engine) applyCondition(query *gorm.DB, field, operator string, value interface{}) *gorm.DB {
	if operator == "" {
		operator = "="
	}

	switch strings.ToLower(operator) {
	case "=", "==":
		return query.Where(field+" = ?", value)
	case "!=", "<>":
		return query.Where(field+" != ?", value)
	case ">":
		return query.Where(field+" > ?", value)
	case "<":
		return query.Where(field+" < ?", value)
	case ">=":
		return query.Where(field+" >= ?", value)
	case "<=":
		return query.Where(field+" <= ?", value)
	case "in":
		return query.Where(field+" IN ?", value)
	case "not in":
		return query.Where(field+" NOT IN ?", value)
	case "like":
		return query.Where(field+" LIKE ?", value)
	case "is null":
		return query.Where(field + " IS NULL")
	case "is not null":
		return query.Where(field + " IS NOT NULL")
	default:
		return query.Where(field+" = ?", value)
	}
}

// resolveValue 解析值（支持表达式）
func (e *Engine) resolveValue(value interface{}, valueFrom string, ctx *Context) interface{} {
	// 如果指定了 valueFrom，从上下文获取
	if valueFrom != "" {
		return e.getValueFromContext(valueFrom, ctx)
	}

	// 如果是字符串，检查是否是表达式
	if str, ok := value.(string); ok {
		if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
			// 表达式 ${xxx}
			expr := strings.Trim(str, "${} ")
			return e.evaluateExpression(expr, ctx)
		}
	}

	return value
}

// getValueFromContext 从上下文获取值
func (e *Engine) getValueFromContext(path string, ctx *Context) interface{} {
	parts := strings.Split(path, ".")

	switch parts[0] {
	case "params":
		if len(parts) > 1 {
			return getNestedValue(ctx.Params, parts[1:])
		}
	case "user", "userId", "userID":
		return ctx.UserID
	case "results":
		if len(parts) > 1 {
			return getNestedValue(ctx.Results, parts[1:])
		}
	case "insertedIds", "insertedIDs":
		if len(parts) > 1 {
			if id, ok := ctx.InsertedIDs[parts[1]]; ok {
				return id
			}
		}
	}

	return nil
}

// evaluateExpression 计算表达式
func (e *Engine) evaluateExpression(expr string, ctx *Context) interface{} {
	// 简单的表达式解析
	// 支持: ${alias.field}, ${params.xxx}, ${user.id}
	// 支持: ${alias.field - value}, ${alias.field + value}

	// 先尝试直接获取值
	if !strings.ContainsAny(expr, "+-*/") {
		return e.getValueFromContext(expr, ctx)
	}

	// 解析算术表达式
	re := regexp.MustCompile(`^(\w+\.?\w*)\s*([+\-*/])\s*(\d+\.?\d*)$`)
	matches := re.FindStringSubmatch(expr)

	if len(matches) == 4 {
		left := e.getValueFromContext(matches[1], ctx)
		op := matches[2]
		right, _ := strconv.ParseFloat(matches[3], 64)

		leftFloat := toFloat64(left)

		switch op {
		case "+":
			return leftFloat + right
		case "-":
			return leftFloat - right
		case "*":
			return leftFloat * right
		case "/":
			if right != 0 {
				return leftFloat / right
			}
		}
	}

	return nil
}

// evaluateCondition 计算条件
func (e *Engine) evaluateCondition(condition string, ctx *Context) bool {
	// 支持的条件格式:
	// ${wallet.balance} >= ${order.totalAmount}
	// ${order.status} == "pending"

	// 解析比较表达式
	re := regexp.MustCompile(`\$\{([^}]+)\}\s*(>=|<=|>|<|==|!=)\s*(\$\{[^}]+\}|\d+\.?\d*|"[^"]*"|'[^']*')`)
	matches := re.FindStringSubmatch(condition)

	if len(matches) == 4 {
		leftExpr := matches[1]
		op := matches[2]
		rightStr := matches[3]

		left := e.getValueFromContext(leftExpr, ctx)

		var right interface{}
		if strings.HasPrefix(rightStr, "${") {
			rightExpr := strings.Trim(rightStr, "${} ")
			right = e.getValueFromContext(rightExpr, ctx)
		} else if strings.HasPrefix(rightStr, `"`) || strings.HasPrefix(rightStr, `'`) {
			right = strings.Trim(rightStr, `"'`)
		} else {
			right, _ = strconv.ParseFloat(rightStr, 64)
		}

		return compareValues(left, op, right)
	}

	return false
}

// TransactionResult 事务执行结果
type TransactionResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Context 事务执行上下文
type Context struct {
	Params      map[string]interface{} `json:"params"`
	UserID      uint                   `json:"userId"`
	Results     map[string]interface{} `json:"results"`
	InsertedIDs map[string]uint        `json:"insertedIds"`
}

// getNestedValue 获取嵌套值
func getNestedValue(data map[string]interface{}, keys []string) interface{} {
	var current interface{} = data
	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		default:
			return nil
		}
	}
	return current
}

// toFloat64 转换为 float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint64:
		return float64(val)
	default:
		return 0
	}
}

// compareValues 比较值
func compareValues(left interface{}, op string, right interface{}) bool {
	leftFloat := toFloat64(left)
	rightFloat := toFloat64(right)

	// 如果是字符串比较
	leftStr, leftIsStr := left.(string)
	rightStr, rightIsStr := right.(string)

	switch op {
	case "==", "=":
		if leftIsStr && rightIsStr {
			return leftStr == rightStr
		}
		return reflect.DeepEqual(left, right)
	case "!=":
		if leftIsStr && rightIsStr {
			return leftStr != rightStr
		}
		return !reflect.DeepEqual(left, right)
	case ">":
		return leftFloat > rightFloat
	case "<":
		return leftFloat < rightFloat
	case ">=":
		return leftFloat >= rightFloat
	case "<=":
		return leftFloat <= rightFloat
	}
	return false
}
