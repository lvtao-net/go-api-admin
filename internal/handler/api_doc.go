package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
	"github.com/lvtao/go-gin-api-admin/pkg/validator"
)

// API Doc Handler
type APIDocHandler struct {
	collectionService *service.CollectionService
}

func NewAPIDocHandler() *APIDocHandler {
	return &APIDocHandler{
		collectionService: service.NewCollectionService(),
	}
}

// GetDoc 获取API文档
func (h *APIDocHandler) GetDoc(c *gin.Context) {
	collections, _, err := h.collectionService.List(1, 100)
	if err != nil {
		response.Error(c, 500, "Failed to get collections")
		return
	}

	// 生成API文档
	docs := generateAPIDoc(collections)

	response.Success(c, docs)
}

// APIInfo API信息
type APIInfo struct {
	Title       string            `json:"title"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	BaseURL     string            `json:"baseURL"`
	Schemes     []string          `json:"schemes"`
	Collections []CollectionDoc   `json:"collections"`
	Endpoints   []Endpoint        `json:"endpoints"`
}

// CollectionDoc 集合文档
type CollectionDoc struct {
	Name        string       `json:"name"`
	Label       string       `json:"label,omitempty"`
	Type        string       `json:"type"`
	Description string       `json:"description,omitempty"`
	Fields      []FieldDoc   `json:"fields"`
	// API 规则
	ListRule   *string `json:"listRule,omitempty"`
	ViewRule   *string `json:"viewRule,omitempty"`
	CreateRule *string `json:"createRule,omitempty"`
	UpdateRule *string `json:"updateRule,omitempty"`
	DeleteRule *string `json:"deleteRule,omitempty"`
}

// FieldDoc 字段文档
type FieldDoc struct {
	Name            string                 `json:"name"`
	Label           string                 `json:"label,omitempty"`
	Type            string                 `json:"type"`
	Required        bool                   `json:"required"`
	Unique          bool                   `json:"unique,omitempty"`
	DefaultValue    interface{}            `json:"defaultValue,omitempty"`
	Description     string                 `json:"description,omitempty"`
	ValidationRules []ValidationRuleDoc    `json:"validationRules,omitempty"`
	Options         map[string]interface{} `json:"options,omitempty"`
	RelationInfo    *RelationInfo          `json:"relationInfo,omitempty"`
}

// ValidationRuleDoc 验证规则文档
type ValidationRuleDoc struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Params      string `json:"params,omitempty"` // 参数说明
}

// RelationInfo 关联信息
type RelationInfo struct {
	Collection   string `json:"collection"`
	LabelField   string `json:"labelField,omitempty"`
	Max          int    `json:"max,omitempty"`
}

// Endpoint 端点
type Endpoint struct {
	Path         string        `json:"path"`
	Method       string        `json:"method"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description"`
	Collection   string        `json:"collection"`
	Auth         string        `json:"auth"`
	Parameters   []Param      `json:"parameters,omitempty"`
	RequestBody  *RequestBody `json:"requestBody,omitempty"`
	Responses    map[string]Response `json:"responses"`
}

// Param 参数
type Param struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Example     string `json:"example,omitempty"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"`
}

// Content 内容
type Content struct {
	Schema  interface{} `json:"schema"`
	Example interface{} `json:"example,omitempty"`
}

// Response 响应
type Response struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content,omitempty"`
}

func generateAPIDoc(collections []model.Collection) APIInfo {
	endpoints := []Endpoint{}
	collectionDocs := make([]CollectionDoc, 0, len(collections))

	// 遍历所有集合，生成对应的API端点
	for _, col := range collections {
		// 生成集合文档
		collectionDocs = append(collectionDocs, generateCollectionDoc(col))

		basePath := "/api/collections/" + col.Name

		// View 类型：只有查询端点
		if col.Type == "view" {
			if !isRuleDisabled(col.ListRule) {
				// 构建参数列表
				params := []Param{
					{Name: "page", In: "query", Description: "页码", Required: false, Type: "integer", Example: "1"},
					{Name: "perPage", In: "query", Description: "每页数量", Required: false, Type: "integer", Example: "30"},
					{Name: "filter", In: "query", Description: "过滤条件", Required: false, Type: "string", Example: "status='active'"},
				}
				// 添加自定义路由参数
				for _, rp := range col.RouteParams {
					if rp.Source == "query" {
						params = append(params, Param{
							Name:        rp.Name,
							In:          "query",
							Description: rp.Description,
							Required:    rp.Required,
							Type:        rp.Type,
							Example:     rp.Default,
						})
					}
				}

				endpoints = append(endpoints, Endpoint{
					Path:        basePath,
					Method:      "GET",
					Summary:     "查询" + col.Label,
					Description: "执行视图查询，返回数据列表",
					Collection:  col.Name,
					Auth:        getAuthNote(col.ListRule),
					Parameters:  params,
					Responses: map[string]Response{
						"200": {Description: "成功"},
					},
				})
			}
			continue
		}

		// Transaction 类型：只有执行端点
		if col.Type == "transaction" {
			if !isRuleDisabled(col.CreateRule) {
				// 构建事务参数示例
				paramsExample := make(map[string]interface{})
				for _, rp := range col.RouteParams {
					if rp.Source == "body" {
						paramsExample[rp.Name] = getExampleValueForType(rp.Type, rp.Default)
					}
				}
				if len(paramsExample) == 0 {
					paramsExample["orderId"] = "123"
					paramsExample["amount"] = 100
				}

				endpoints = append(endpoints, Endpoint{
					Path:        basePath,
					Method:      "POST",
					Summary:     "执行" + col.Label,
					Description: "执行事务操作",
					Collection:  col.Name,
					Auth:        getAuthNote(col.CreateRule),
					RequestBody: &RequestBody{
						Description: "事务参数",
						Content: map[string]Content{
							"application/json": {
								Schema:  map[string]interface{}{"type": "object"},
								Example: map[string]interface{}{"params": paramsExample},
							},
						},
					},
					Responses: map[string]Response{
						"200": {Description: "执行成功"},
						"400": {Description: "执行失败"},
					},
				})
			}
			continue
		}

		// Base/Auth 类型：完整的 CRUD 端点

		// 列表 - 跳过禁用的API
		if !isRuleDisabled(col.ListRule) {
			endpoints = append(endpoints, Endpoint{
				Path:        basePath + "/records",
				Method:      "GET",
				Summary:     "获取" + col.Name + "列表",
				Description: "获取" + col.Name + "的所有记录，支持分页和过滤",
				Collection:  col.Name,
				Auth:        getAuthNote(col.ListRule),
				Parameters: []Param{
					{Name: "page", In: "query", Description: "页码", Required: false, Type: "integer", Example: "1"},
					{Name: "perPage", In: "query", Description: "每页数量", Required: false, Type: "integer", Example: "30"},
					{Name: "sort", In: "query", Description: "排序字段，-前缀表示倒序", Required: false, Type: "string", Example: "-created"},
					{Name: "filter", In: "query", Description: "过滤条件", Required: false, Type: "string", Example: "status='active'"},
				},
				Responses: map[string]Response{
					"200": {
						Description: "成功",
						Content: map[string]Content{
							"application/json": {
								Schema: map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"code":    map[string]string{"type": "integer"},
										"message": map[string]string{"type": "string"},
										"data": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"items":      map[string]string{"type": "array"},
												"page":       map[string]string{"type": "integer"},
												"perPage":    map[string]string{"type": "integer"},
												"totalItems": map[string]string{"type": "integer"},
											},
										},
									},
								},
							},
						},
					},
				},
			})
		}

		// 单条记录 - 跳过禁用的API
		if !isRuleDisabled(col.ViewRule) {
			endpoints = append(endpoints, Endpoint{
				Path:        basePath + "/records/:id",
				Method:      "GET",
				Summary:     "获取" + col.Name + "单条记录",
				Description: "根据ID获取" + col.Name + "的详细信息",
				Collection:  col.Name,
				Auth:        getAuthNote(col.ViewRule),
				Parameters: []Param{
					{Name: "id", In: "path", Description: "记录ID", Required: true, Type: "string", Example: "123"},
				},
				Responses: map[string]Response{
					"200": {Description: "成功"},
					"404": {Description: "记录不存在"},
				},
			})
		}

		// 创建 - 跳过禁用的API
		if !isRuleDisabled(col.CreateRule) {
			endpoints = append(endpoints, Endpoint{
				Path:        basePath + "/records",
				Method:      "POST",
				Summary:     "创建" + col.Name + "记录",
				Description: "创建一条新的" + col.Name + "记录",
				Collection:  col.Name,
				Auth:        getAuthNote(col.CreateRule),
				RequestBody: &RequestBody{
					Description: "要创建的记录数据",
					Content: map[string]Content{
						"application/json": {
							Schema:  map[string]interface{}{"type": "object"},
							Example: generateExample(col.Fields),
						},
					},
				},
				Responses: map[string]Response{
					"201": {Description: "创建成功"},
					"400": {Description: "请求参数错误"},
				},
			})
		}

		// 更新 - 跳过禁用的API
		if !isRuleDisabled(col.UpdateRule) {
			endpoints = append(endpoints, Endpoint{
				Path:        basePath + "/records/:id",
				Method:      "PATCH",
				Summary:     "更新" + col.Name + "记录",
				Description: "更新指定ID的记录，只传递需要更新的字段",
				Collection:  col.Name,
				Auth:        getAuthNote(col.UpdateRule),
				Parameters: []Param{
					{Name: "id", In: "path", Description: "记录ID", Required: true, Type: "string", Example: "123"},
				},
				RequestBody: &RequestBody{
					Description: "要更新的字段数据",
					Content: map[string]Content{
						"application/json": {
							Schema:  map[string]interface{}{"type": "object"},
							Example: generateUpdateExample(col.Fields),
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "更新成功"},
					"404": {Description: "记录不存在"},
				},
			})
		}

		// 删除 - 跳过禁用的API
		if !isRuleDisabled(col.DeleteRule) {
			endpoints = append(endpoints, Endpoint{
				Path:        basePath + "/records/:id",
				Method:      "DELETE",
				Summary:     "删除" + col.Name + "记录",
				Description: "删除指定ID的记录",
				Collection:  col.Name,
				Auth:        getAuthNote(col.DeleteRule),
				Parameters: []Param{
					{Name: "id", In: "path", Description: "记录ID", Required: true, Type: "string"},
				},
				Responses: map[string]Response{
					"200": {Description: "删除成功"},
					"404": {Description: "记录不存在"},
				},
			})
		}

		// Auth 集合特殊端点
		if col.Type == "auth" {
			authPath := basePath

			// 注册
			endpoints = append(endpoints, Endpoint{
				Path:        authPath + "/register",
				Method:      "POST",
				Summary:     col.Name + " 用户注册",
				Description: "注册新用户。identity 自动判断类型：带@为邮箱(email)，手机号为mobile，其他为account。如需验证，先调用 /request-otp 获取验证码",
				Collection:  col.Name,
				Auth:        "公开",
				Parameters: []Param{
					{Name: "identity", In: "body", Description: "邮箱/手机号/账号", Required: true, Type: "string"},
					{Name: "password", In: "body", Description: "登录密码", Required: true, Type: "string"},
					{Name: "code", In: "body", Description: "验证码（可选，如需验证先调用 /request-otp）", Required: false, Type: "string"},
				},
				RequestBody: &RequestBody{
					Description: "注册信息",
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"identity": map[string]string{"type": "string", "description": "邮箱/手机号/账号"},
									"password": map[string]string{"type": "string", "description": "登录密码"},
									"code":     map[string]string{"type": "string", "description": "验证码（可选）"},
								},
								"required": []string{"identity", "password"},
							},
							Example: map[string]string{"identity": "user@example.com", "password": "your_password", "code": "123456"},
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "注册成功"},
					"400": {Description: "参数错误或用户已存在"},
				},
			})

			// 登录
			endpoints = append(endpoints, Endpoint{
				Path:        authPath + "/auth-with-password",
				Method:      "POST",
				Summary:     col.Name + " 用户登录",
				Description: "使用账号密码登录。identity 自动判断类型：带@为邮箱(email)，手机号为mobile，其他为account",
				Collection:  col.Name,
				Auth:        "公开",
				Parameters: []Param{
					{Name: "identity", In: "body", Description: "邮箱/手机号/账号", Required: true, Type: "string"},
					{Name: "password", In: "body", Description: "登录密码", Required: true, Type: "string"},
				},
				RequestBody: &RequestBody{
					Description: "登录凭证",
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"identity": map[string]string{"type": "string", "description": "邮箱/手机号/账号"},
									"password": map[string]string{"type": "string", "description": "登录密码"},
								},
								"required": []string{"identity", "password"},
							},
							Example: map[string]string{"identity": "user@example.com", "password": "your_password"},
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "登录成功，返回token"},
					"400": {Description: "账号或密码错误"},
				},
			})

			// 刷新Token
			endpoints = append(endpoints, Endpoint{
				Path:        authPath + "/auth-refresh",
				Method:      "POST",
				Summary:     col.Name + " 刷新Token",
				Description: "使用refresh token刷新访问令牌",
				Collection:  col.Name,
				Auth:        "公开",
				Parameters: []Param{
					{Name: "refreshToken", In: "body", Description: "刷新令牌", Required: true, Type: "string"},
				},
				RequestBody: &RequestBody{
					Description: "刷新令牌",
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"refreshToken": map[string]string{"type": "string", "description": "刷新令牌"},
								},
								"required": []string{"refreshToken"},
							},
							Example: map[string]string{"refreshToken": "your_refresh_token"},
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "刷新成功，返回新token"},
					"400": {Description: "无效的refresh token"},
				},
			})

			// 请求验证码
			endpoints = append(endpoints, Endpoint{
				Path:        authPath + "/request-otp",
				Method:      "POST",
				Summary:     col.Name + " 请求验证码",
				Description: "请求验证码。identity 自动判断类型：带@为邮箱，手机号为mobile，其他为account。type: register-注册验证, password-reset-找回密码",
				Collection:  col.Name,
				Auth:        "公开",
				Parameters: []Param{
					{Name: "identity", In: "body", Description: "邮箱/手机号/账号", Required: true, Type: "string"},
					{Name: "type", In: "body", Description: "类型：register/password-reset", Required: true, Type: "string"},
				},
				RequestBody: &RequestBody{
					Description: "验证码请求",
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"identity": map[string]string{"type": "string", "description": "邮箱/手机号/账号"},
									"type":     map[string]string{"type": "string", "description": "类型：register/password-reset"},
								},
								"required": []string{"identity", "type"},
							},
							Example: map[string]string{"identity": "user@example.com", "type": "register"},
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "验证码已发送"},
					"400": {Description: "参数错误"},
				},
			})

			// 重置密码
			endpoints = append(endpoints, Endpoint{
				Path:        authPath + "/reset-password",
				Method:      "POST",
				Summary:     col.Name + " 重置密码",
				Description: "使用验证码重置密码。identity 自动判断类型：带@为邮箱，手机号为mobile，其他为account",
				Collection:  col.Name,
				Auth:        "公开",
				Parameters: []Param{
					{Name: "identity", In: "body", Description: "邮箱/手机号/账号", Required: true, Type: "string"},
					{Name: "code", In: "body", Description: "验证码", Required: true, Type: "string"},
					{Name: "password", In: "body", Description: "新密码", Required: true, Type: "string"},
				},
				RequestBody: &RequestBody{
					Description: "重置密码请求",
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"identity": map[string]string{"type": "string", "description": "邮箱/手机号/账号"},
									"code":     map[string]string{"type": "string", "description": "验证码"},
									"password": map[string]string{"type": "string", "description": "新密码"},
								},
								"required": []string{"identity", "code", "password"},
							},
							Example: map[string]string{"identity": "user@example.com", "code": "123456", "password": "new_password"},
						},
					},
				},
				Responses: map[string]Response{
					"200": {Description: "密码重置成功"},
					"400": {Description: "验证码无效或已过期"},
				},
			})
		}
	}

	// 添加公共接口
	// 文件上传
	endpoints = append(endpoints, Endpoint{
		Path:        "/api/files/upload",
		Method:      "POST",
		Summary:     "文件上传",
		Description: "上传单个文件，返回文件路径。字段名必须为 file，支持常见文件类型",
		Collection:  "_public",
		Auth:        "公开",
		RequestBody: &RequestBody{
			Content: map[string]Content{
				"multipart/form-data": {
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"file": map[string]string{"type": "file", "description": "要上传的文件（字段名必须为 file）"},
						},
					},
				},
			},
		},
		Responses: map[string]Response{
			"200": {Description: "上传成功，返回文件路径"},
			"400": {Description: "上传失败"},
		},
	})

	return APIInfo{
		Title:       "Go Gin API Admin",
		Version:     "1.0.0",
		Description: "基于集合配置的动态API管理后台",
		BaseURL:     "/api",
		Schemes:     []string{"http", "https"},
		Collections: collectionDocs,
		Endpoints:   endpoints,
	}
}

func isRuleDisabled(rule *string) bool {
	if rule == nil {
		return false
	}
	return *rule == "false"
}

func getAuthNote(rule *string) string {
	if rule == nil || *rule == "" {
		return "公开"
	}
	if *rule == "false" {
		return "已禁用"
	}
	return "需要认证"
}

// generateCollectionDoc 生成集合文档
func generateCollectionDoc(col model.Collection) CollectionDoc {
	fields := make([]FieldDoc, 0, len(col.Fields))

	for _, field := range col.Fields {
		// 跳过 API 禁用的字段
		if field.APIDisabled {
			continue
		}

		fieldDoc := FieldDoc{
			Name:         field.Name,
			Label:        field.Label,
			Type:         field.Type,
			Required:     field.Required,
			Unique:       field.Unique,
			DefaultValue: field.DefaultValue,
			Description:  field.Description,
			Options:      field.Options,
		}

		// 添加验证规则详细说明
		if len(field.ValidationRules) > 0 {
			fieldDoc.ValidationRules = make([]ValidationRuleDoc, 0, len(field.ValidationRules))
			for _, ruleName := range field.ValidationRules {
				ruleDoc := getValidationRuleDoc(ruleName)
				fieldDoc.ValidationRules = append(fieldDoc.ValidationRules, ruleDoc)
			}
		}

		// 如果是必填字段，添加必填规则
		if field.Required {
			hasRequired := false
			for _, r := range fieldDoc.ValidationRules {
				if r.Name == "required" {
					hasRequired = true
					break
				}
			}
			if !hasRequired {
				fieldDoc.ValidationRules = append([]ValidationRuleDoc{{
					Name:        "required",
					Label:       "必填",
					Description: "字段不能为空",
				}}, fieldDoc.ValidationRules...)
			}
		}

		// 添加关联信息
		if field.Type == "relation" && field.RelationCollection != "" {
			fieldDoc.RelationInfo = &RelationInfo{
				Collection: field.RelationCollection,
				LabelField: field.RelationLabelField,
				Max:        field.RelationMax,
			}
		}

		fields = append(fields, fieldDoc)
	}

	return CollectionDoc{
		Name:        col.Name,
		Label:       col.Label,
		Type:        col.Type,
		Description: col.Description,
		Fields:      fields,
		// API 规则
		ListRule:   col.ListRule,
		ViewRule:   col.ViewRule,
		CreateRule: col.CreateRule,
		UpdateRule: col.UpdateRule,
		DeleteRule: col.DeleteRule,
	}
}

// getValidationRuleDoc 获取验证规则文档
func getValidationRuleDoc(ruleStr string) ValidationRuleDoc {
	// 解析规则名称（去掉参数部分）
	ruleName := ruleStr
	if idx := strings.Index(ruleStr, ":"); idx != -1 {
		ruleName = ruleStr[:idx]
	}

	// 从内置规则中查找
	for _, rule := range validator.BuiltInRules {
		if rule.Name == ruleName {
			// 解析参数值
			params := parseRuleParams(ruleStr)
			description := rule.Description
			
			// 如果有参数，添加到描述中
			if len(params) > 0 {
				paramStr := ""
				for k, v := range params {
					if paramStr != "" {
						paramStr += ", "
					}
					paramStr += fmt.Sprintf("%s=%v", k, v)
				}
				description = fmt.Sprintf("%s [%s]", rule.Description, paramStr)
			}
			
			return ValidationRuleDoc{
				Name:        rule.Name,
				Label:       rule.Label,
				Description: description,
				Params:      formatRuleParams(rule.Params),
			}
		}
	}

	// 未找到，返回原始名称
	return ValidationRuleDoc{
		Name:        ruleName,
		Label:       ruleName,
		Description: "自定义验证规则",
	}
}

// parseRuleParams 解析规则参数
func parseRuleParams(ruleStr string) map[string]string {
	params := make(map[string]string)
	
	// 检查是否有参数，格式: rule_name:param1=value1,param2=value2
	if idx := strings.Index(ruleStr, ":"); idx != -1 {
		paramStr := ruleStr[idx+1:]
		
		for _, pair := range strings.Split(paramStr, ",") {
			if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	
	return params
}

// formatRuleParams 格式化规则参数说明
func formatRuleParams(params []validator.RuleParam) string {
	if len(params) == 0 {
		return ""
	}

	result := ""
	for i, p := range params {
		if i > 0 {
			result += ", "
		}
		result += p.Label
		if p.Default != nil {
			result += " (默认: " + formatParamValue(p.Default) + ")"
		}
	}
	return result
}

// formatParamValue 格式化参数值
func formatParamValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func generateExample(fields []model.CollectionField) map[string]interface{} {
	example := make(map[string]interface{})
	for _, field := range fields {
		switch field.Type {
		case "text":
			example[field.Name] = "sample text"
		case "number":
			example[field.Name] = 123
		case "bool":
			example[field.Name] = true
		case "email":
			example[field.Name] = "user@example.com"
		case "date":
			example[field.Name] = "2024-01-01"
		case "json":
			example[field.Name] = map[string]interface{}{"key": "value"}
		default:
			example[field.Name] = nil
		}
	}
	return example
}

// generateUpdateExample 生成更新示例（只包含前3个字段）
func generateUpdateExample(fields []model.CollectionField) map[string]interface{} {
	example := make(map[string]interface{})
	count := 0
	for _, field := range fields {
		if count >= 3 {
			break
		}
		// 跳过系统字段
		if field.Name == "id" || field.Name == "created" || field.Name == "updated" {
			continue
		}
		switch field.Type {
		case "text":
			example[field.Name] = "updated text"
		case "number":
			example[field.Name] = 456
		case "bool":
			example[field.Name] = false
		case "email":
			example[field.Name] = "new@example.com"
		case "date":
			example[field.Name] = "2024-06-01"
		default:
			example[field.Name] = nil
		}
		count++
	}
	return example
}

// getExampleValueForType 根据类型获取示例值
func getExampleValueForType(typeName string, defaultVal string) interface{} {
	if defaultVal != "" {
		return defaultVal
	}
	switch typeName {
	case "number", "int", "integer", "float":
		return 100
	case "bool", "boolean":
		return true
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return "example"
	}
}
