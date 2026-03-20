package validator

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// 内置验证规则常量
const (
	RuleRequired         = "required"         // 必填
	RuleEmail            = "email"            // 邮箱
	RulePhone            = "phone"            // 手机号（中国）
	RuleURL              = "url"              // URL
	RuleIDCard           = "idcard"           // 身份证号（中国）
	RuleIP               = "ip"               // IP地址（v4或v6）
	RuleIPv4             = "ipv4"             // IPv4
	RuleIPv6             = "ipv6"             // IPv6
	RuleNumber           = "number"           // 数字
	RuleInteger          = "integer"          // 整数
	RulePositive         = "positive"         // 正数
	RuleNegative         = "negative"         // 负数
	RuleAlpha            = "alpha"            // 纯字母
	RuleAlphaNum         = "alphanum"         // 字母和数字
	RuleChinese          = "chinese"          // 中文字符
	RuleDate             = "date"             // 日期格式
	RuleDateTime         = "datetime"         // 日期时间格式
	RuleMinLength        = "min_length"       // 最小长度
	RuleMaxLength        = "max_length"       // 最大长度
	RuleRangeLength      = "range_length"     // 长度范围
	RuleMinValue         = "min_value"        // 最小值
	RuleMaxValue         = "max_value"        // 最大值
	RuleRangeValue       = "range_value"      // 值范围
	RulePasswordStrength = "password_strength" // 密码强度
	RuleCreditCard       = "credit_card"      // 信用卡号
	RuleWechat           = "wechat"           // 微信号
	RuleQQ               = "qq"               // QQ号
	RuleBankCard         = "bank_card"        // 银行卡号
	RuleHexColor         = "hex_color"        // 十六进制颜色
	RuleJSON             = "json"             // JSON格式
	RuleUUID             = "uuid"             // UUID
	RuleNoSpace          = "no_space"         // 不含空格
	RuleNoSpecialChar    = "no_special_char"  // 不含特殊字符
)

// RuleDefinition 规则定义
type RuleDefinition struct {
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Params      []RuleParam       `json:"params,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
}

// RuleParam 规则参数
type RuleParam struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"` // number, string, select
	Default     interface{} `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Description string `json:"description,omitempty"`
}

// 内置规则定义
var BuiltInRules = []RuleDefinition{
	{Name: RuleRequired, Label: "必填", Description: "字段不能为空", Category: "基础"},
	{Name: RuleEmail, Label: "邮箱", Description: "有效的邮箱地址", Category: "格式"},
	{Name: RulePhone, Label: "手机号", Description: "中国大陆手机号码", Category: "格式"},
	{Name: RuleURL, Label: "URL", Description: "有效的URL地址", Category: "格式"},
	{Name: RuleIDCard, Label: "身份证号", Description: "中国大陆身份证号码", Category: "格式"},
	{Name: RuleIP, Label: "IP地址", Description: "IPv4或IPv6地址", Category: "格式"},
	{Name: RuleIPv4, Label: "IPv4", Description: "IPv4地址", Category: "格式"},
	{Name: RuleIPv6, Label: "IPv6", Description: "IPv6地址", Category: "格式"},
	{Name: RuleNumber, Label: "数字", Description: "有效的数字", Category: "类型"},
	{Name: RuleInteger, Label: "整数", Description: "整数", Category: "类型"},
	{Name: RulePositive, Label: "正数", Description: "大于0的数", Category: "类型"},
	{Name: RuleNegative, Label: "负数", Description: "小于0的数", Category: "类型"},
	{Name: RuleAlpha, Label: "纯字母", Description: "只包含字母", Category: "格式"},
	{Name: RuleAlphaNum, Label: "字母数字", Description: "只包含字母和数字", Category: "格式"},
	{Name: RuleChinese, Label: "中文", Description: "只包含中文字符", Category: "格式"},
	{Name: RuleDate, Label: "日期", Description: "日期格式 YYYY-MM-DD", Category: "格式"},
	{Name: RuleDateTime, Label: "日期时间", Description: "日期时间格式", Category: "格式"},
	{Name: RuleMinLength, Label: "最小长度", Description: "最小字符长度", Category: "长度", Params: []RuleParam{{Name: "min", Label: "最小长度", Type: "number", Default: 1}}},
	{Name: RuleMaxLength, Label: "最大长度", Description: "最大字符长度", Category: "长度", Params: []RuleParam{{Name: "max", Label: "最大长度", Type: "number", Default: 255}}},
	{Name: RuleRangeLength, Label: "长度范围", Description: "字符长度范围", Category: "长度", Params: []RuleParam{{Name: "min", Label: "最小长度", Type: "number", Default: 1}, {Name: "max", Label: "最大长度", Type: "number", Default: 255}}},
	{Name: RuleMinValue, Label: "最小值", Description: "数字最小值", Category: "范围", Params: []RuleParam{{Name: "min", Label: "最小值", Type: "number", Default: 0}}},
	{Name: RuleMaxValue, Label: "最大值", Description: "数字最大值", Category: "范围", Params: []RuleParam{{Name: "max", Label: "最大值", Type: "number", Default: 100}}},
	{Name: RuleRangeValue, Label: "值范围", Description: "数字值范围", Category: "范围", Params: []RuleParam{{Name: "min", Label: "最小值", Type: "number", Default: 0}, {Name: "max", Label: "最大值", Type: "number", Default: 100}}},
	{Name: RulePasswordStrength, Label: "密码强度", Description: "密码强度检查[至少8位，包含大小写字母和数字]", Category: "安全"},
	{Name: RuleCreditCard, Label: "信用卡号", Description: "有效的信用卡号", Category: "格式"},
	{Name: RuleWechat, Label: "微信号", Description: "微信号格式", Category: "格式"},
	{Name: RuleQQ, Label: "QQ号", Description: "QQ号码", Category: "格式"},
	{Name: RuleBankCard, Label: "银行卡号", Description: "银行卡号格式", Category: "格式"},
	{Name: RuleHexColor, Label: "十六进制颜色", Description: "如 #FFF 或 #FFFFFF", Category: "格式"},
	{Name: RuleJSON, Label: "JSON格式", Description: "有效的JSON格式", Category: "格式"},
	{Name: RuleUUID, Label: "UUID", Description: "UUID格式", Category: "格式"},
	{Name: RuleNoSpace, Label: "不含空格", Description: "不能包含空格字符", Category: "格式"},
	{Name: RuleNoSpecialChar, Label: "不含特殊字符", Description: "不能包含特殊字符", Category: "格式"},
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// Validator 验证器
type Validator struct {
	value   interface{}
	field   string
	rules   []string
	params  map[string]interface{}
	customMessages map[string]string
}

// NewValidator 创建验证器
func NewValidator(field string, value interface{}, rules []string, customMessages map[string]string) *Validator {
	return &Validator{
		field:   field,
		value:   value,
		rules:   rules,
		params:  make(map[string]interface{}),
		customMessages: customMessages,
	}
}

// SetParams 设置规则参数
func (v *Validator) SetParams(params map[string]interface{}) {
	v.params = params
}

// Validate 执行验证
func (v *Validator) Validate() []ValidationError {
	var errors []ValidationError

	for _, rule := range v.rules {
		if err := v.validateRule(rule); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateRule 验证单个规则
func (v *Validator) validateRule(rule string) *ValidationError {
	// 解析规则名称和参数
	ruleName, params := parseRule(rule)
	
	// 合并参数
	mergedParams := make(map[string]interface{})
	for k, val := range v.params {
		mergedParams[k] = val
	}
	for k, val := range params {
		mergedParams[k] = val
	}

	var valid bool
	var defaultMessage string

	switch ruleName {
	case RuleRequired:
		valid = validateRequired(v.value)
		defaultMessage = "字段不能为空"
	case RuleEmail:
		valid = validateEmail(v.value)
		defaultMessage = "请输入有效的邮箱地址"
	case RulePhone:
		valid = validatePhone(v.value)
		defaultMessage = "请输入有效的手机号码"
	case RuleURL:
		valid = validateURL(v.value)
		defaultMessage = "请输入有效的URL地址"
	case RuleIDCard:
		valid = validateIDCard(v.value)
		defaultMessage = "请输入有效的身份证号码"
	case RuleIP:
		valid = validateIP(v.value)
		defaultMessage = "请输入有效的IP地址"
	case RuleIPv4:
		valid = validateIPv4(v.value)
		defaultMessage = "请输入有效的IPv4地址"
	case RuleIPv6:
		valid = validateIPv6(v.value)
		defaultMessage = "请输入有效的IPv6地址"
	case RuleNumber:
		valid = validateNumber(v.value)
		defaultMessage = "请输入有效的数字"
	case RuleInteger:
		valid = validateInteger(v.value)
		defaultMessage = "请输入整数"
	case RulePositive:
		valid = validatePositive(v.value)
		defaultMessage = "请输入正数"
	case RuleNegative:
		valid = validateNegative(v.value)
		defaultMessage = "请输入负数"
	case RuleAlpha:
		valid = validateAlpha(v.value)
		defaultMessage = "只能包含字母"
	case RuleAlphaNum:
		valid = validateAlphaNum(v.value)
		defaultMessage = "只能包含字母和数字"
	case RuleChinese:
		valid = validateChinese(v.value)
		defaultMessage = "只能包含中文字符"
	case RuleDate:
		valid = validateDate(v.value)
		defaultMessage = "请输入有效的日期格式"
	case RuleDateTime:
		valid = validateDateTime(v.value)
		defaultMessage = "请输入有效的日期时间格式"
	case RuleMinLength:
		min := getParamInt(mergedParams, "min", 1)
		valid = validateMinLength(v.value, min)
		defaultMessage = fmt.Sprintf("长度不能小于%d个字符", min)
	case RuleMaxLength:
		max := getParamInt(mergedParams, "max", 255)
		valid = validateMaxLength(v.value, max)
		defaultMessage = fmt.Sprintf("长度不能超过%d个字符", max)
	case RuleRangeLength:
		min := getParamInt(mergedParams, "min", 1)
		max := getParamInt(mergedParams, "max", 255)
		valid = validateRangeLength(v.value, min, max)
		defaultMessage = fmt.Sprintf("长度必须在%d到%d个字符之间", min, max)
	case RuleMinValue:
		min := getParamFloat(mergedParams, "min", 0)
		valid = validateMinValue(v.value, min)
		defaultMessage = fmt.Sprintf("值不能小于%.2f", min)
	case RuleMaxValue:
		max := getParamFloat(mergedParams, "max", 100)
		valid = validateMaxValue(v.value, max)
		defaultMessage = fmt.Sprintf("值不能超过%.2f", max)
	case RuleRangeValue:
		min := getParamFloat(mergedParams, "min", 0)
		max := getParamFloat(mergedParams, "max", 100)
		valid = validateRangeValue(v.value, min, max)
		defaultMessage = fmt.Sprintf("值必须在%.2f到%.2f之间", min, max)
	case RulePasswordStrength:
		valid = validatePasswordStrength(v.value)
		defaultMessage = "密码至少8位，需包含大小写字母和数字"
	case RuleCreditCard:
		valid = validateCreditCard(v.value)
		defaultMessage = "请输入有效的信用卡号"
	case RuleWechat:
		valid = validateWechat(v.value)
		defaultMessage = "请输入有效的微信号"
	case RuleQQ:
		valid = validateQQ(v.value)
		defaultMessage = "请输入有效的QQ号"
	case RuleBankCard:
		valid = validateBankCard(v.value)
		defaultMessage = "请输入有效的银行卡号"
	case RuleHexColor:
		valid = validateHexColor(v.value)
		defaultMessage = "请输入有效的颜色值"
	case RuleJSON:
		valid = validateJSON(v.value)
		defaultMessage = "请输入有效的JSON格式"
	case RuleUUID:
		valid = validateUUID(v.value)
		defaultMessage = "请输入有效的UUID"
	case RuleNoSpace:
		valid = validateNoSpace(v.value)
		defaultMessage = "不能包含空格"
	case RuleNoSpecialChar:
		valid = validateNoSpecialChar(v.value)
		defaultMessage = "不能包含特殊字符"
	default:
		return nil
	}

	if !valid {
		message := defaultMessage
		if customMsg, ok := v.customMessages[ruleName]; ok {
			message = customMsg
		}
		return &ValidationError{
			Field:   v.field,
			Rule:    ruleName,
			Message: message,
		}
	}

	return nil
}

// parseRule 解析规则字符串
func parseRule(rule string) (string, map[string]interface{}) {
	params := make(map[string]interface{})
	
	// 检查是否有参数，格式: rule_name:param1=value1,param2=value2
	if idx := strings.Index(rule, ":"); idx != -1 {
		ruleName := rule[:idx]
		paramStr := rule[idx+1:]
		
		for _, pair := range strings.Split(paramStr, ",") {
			if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
				params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
		return ruleName, params
	}
	
	return rule, params
}

// getParamInt 获取整数参数
func getParamInt(params map[string]interface{}, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

// getParamFloat 获取浮点数参数
func getParamFloat(params map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
		}
	}
	return defaultVal
}

// stringValue 获取字符串值
func stringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

// floatValue 获取浮点数值
func floatValue(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// ========== 验证函数 ==========

func validateRequired(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) != ""
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	default:
		return true
	}
}

func validateEmail(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true // 空值由required处理
	}
	_, err := mail.ParseAddress(str)
	return err == nil
}

func validatePhone(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 中国大陆手机号：1开头，11位数字
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, str)
	return matched
}

func validateURL(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	_, err := url.ParseRequestURI(str)
	return err == nil
}

func validateIDCard(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 15位或18位身份证号
	matched, _ := regexp.MatchString(`^(\d{15}|\d{17}[\dXx])$`, str)
	return matched
}

func validateIP(v interface{}) bool {
	return validateIPv4(v) || validateIPv6(v)
}

func validateIPv4(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`, str)
	return matched
}

func validateIPv6(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 简化的IPv6验证
	matched, _ := regexp.MatchString(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`, str)
	return matched
}

func validateNumber(v interface{}) bool {
	_, ok := floatValue(v)
	return ok || v == nil || stringValue(v) == ""
}

func validateInteger(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^-?\d+$`, str)
	return matched
}

func validatePositive(v interface{}) bool {
	f, ok := floatValue(v)
	if !ok {
		str := stringValue(v)
		if str == "" {
			return true
		}
		return false
	}
	return f > 0
}

func validateNegative(v interface{}) bool {
	f, ok := floatValue(v)
	if !ok {
		str := stringValue(v)
		if str == "" {
			return true
		}
		return false
	}
	return f < 0
}

func validateAlpha(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func validateAlphaNum(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validateChinese(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^[\p{Han}]+$`, str)
	return matched
}

func validateDate(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, str)
	return matched
}

func validateDateTime(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, str)
	return matched
}

func validateMinLength(v interface{}, min int) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	return len([]rune(str)) >= min
}

func validateMaxLength(v interface{}, max int) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	return len([]rune(str)) <= max
}

func validateRangeLength(v interface{}, min, max int) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	length := len([]rune(str))
	return length >= min && length <= max
}

func validateMinValue(v interface{}, min float64) bool {
	f, ok := floatValue(v)
	if !ok {
		return true
	}
	return f >= min
}

func validateMaxValue(v interface{}, max float64) bool {
	f, ok := floatValue(v)
	if !ok {
		return true
	}
	return f <= max
}

func validateRangeValue(v interface{}, min, max float64) bool {
	f, ok := floatValue(v)
	if !ok {
		return true
	}
	return f >= min && f <= max
}

func validatePasswordStrength(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	if len(str) < 8 {
		return false
	}
	
	var hasUpper, hasLower, hasDigit bool
	for _, r := range str {
		if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	
	return hasUpper && hasLower && hasDigit
}

func validateCreditCard(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 简单的信用卡号验证：13-19位数字
	matched, _ := regexp.MatchString(`^\d{13,19}$`, strings.ReplaceAll(str, " ", ""))
	return matched
}

func validateWechat(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 微信号：6-20位，字母开头，可包含字母、数字、下划线、减号
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_-]{5,19}$`, str)
	return matched
}

func validateQQ(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// QQ号：5-11位数字，不以0开头
	matched, _ := regexp.MatchString(`^[1-9]\d{4,10}$`, str)
	return matched
}

func validateBankCard(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	// 银行卡号：16-19位数字
	matched, _ := regexp.MatchString(`^\d{16,19}$`, strings.ReplaceAll(str, " ", ""))
	return matched
}

func validateHexColor(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`, str)
	return matched
}

func validateJSON(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	var js interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

func validateUUID(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	matched, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, str)
	return matched
}

func validateNoSpace(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	return !strings.Contains(str, " ") && !strings.Contains(str, "\t") && !strings.Contains(str, "\n")
}

func validateNoSpecialChar(v interface{}) bool {
	str := stringValue(v)
	if str == "" {
		return true
	}
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r) {
			return false
		}
	}
	return true
}

// FieldValidator 字段验证器接口
type FieldValidator interface {
	GetName() string
	GetRequired() bool
	GetValidationRules() []string
	GetValidationMessages() map[string]string
	GetMin() interface{}
	GetMax() interface{}
}

// ValidateRecord 验证记录数据（创建时使用，验证所有必填字段）
func ValidateRecord(data map[string]interface{}, fields []FieldValidator) []ValidationError {
	return validateRecordInternal(data, fields, false)
}

// ValidateRecordForUpdate 验证记录数据（更新时使用，只验证提交的字段）
func ValidateRecordForUpdate(data map[string]interface{}, fields []FieldValidator) []ValidationError {
	return validateRecordInternal(data, fields, true)
}

// validateRecordInternal 内部验证函数
// isUpdate: 是否为更新操作，更新时只验证提交的字段
func validateRecordInternal(data map[string]interface{}, fields []FieldValidator, isUpdate bool) []ValidationError {
	var errors []ValidationError

	for _, field := range fields {
		fieldName := field.GetName()
		value, exists := data[fieldName]

		// 更新操作时，只验证提交的字段
		if isUpdate && !exists {
			continue
		}

		rules := field.GetValidationRules()
		customMessages := field.GetValidationMessages()

		// 如果设置了必填，添加required规则
		// 更新操作时，只有字段存在且值为空时才检查必填
		if field.GetRequired() {
			if !isUpdate || (exists && isEmptyValue(value)) {
				rules = append([]string{RuleRequired}, rules...)
			}
		}

		if len(rules) > 0 {
			validator := NewValidator(fieldName, value, rules, customMessages)

			// 添加min/max参数
			params := make(map[string]interface{})
			if field.GetMin() != nil {
				params["min"] = field.GetMin()
			}
			if field.GetMax() != nil {
				params["max"] = field.GetMax()
			}
			validator.SetParams(params)

			if errs := validator.Validate(); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}
	}

	return errors
}

// isEmptyValue 检查值是否为空
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// CollectionFieldAdapter 适配 CollectionField 到 FieldValidator 接口
type CollectionFieldAdapter struct {
	Name               string
	Required           bool
	ValidationRules    []string
	ValidationMessages map[string]string
	Min                interface{}
	Max                interface{}
}

func (f CollectionFieldAdapter) GetName() string               { return f.Name }
func (f CollectionFieldAdapter) GetRequired() bool             { return f.Required }
func (f CollectionFieldAdapter) GetValidationRules() []string  { return f.ValidationRules }
func (f CollectionFieldAdapter) GetValidationMessages() map[string]string { return f.ValidationMessages }
func (f CollectionFieldAdapter) GetMin() interface{}           { return f.Min }
func (f CollectionFieldAdapter) GetMax() interface{}           { return f.Max }

// GetBuiltInRules 获取内置规则列表
func GetBuiltInRules() []RuleDefinition {
	return BuiltInRules
}

// GetRuleByCategory 按类别获取规则
func GetRuleByCategory() map[string][]RuleDefinition {
	result := make(map[string][]RuleDefinition)
	for _, rule := range BuiltInRules {
		result[rule.Category] = append(result[rule.Category], rule)
	}
	return result
}
