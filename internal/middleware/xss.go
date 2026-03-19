package middleware

import (
	"bytes"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// XSSConfig XSS 中间件配置
type XSSConfig struct {
	SkipPaths      []string // 跳过 XSS 过滤的路径
	SkipPrefixes   []string // 跳过 XSS 过滤的路径前缀
	SkipQueryKeys  []string // 跳过 XSS 过滤的查询参数名
	StripTags      bool     // 是否移除 HTML 标签
	EscapeHTML     bool     // 是否转义 HTML 字符
}

// DefaultXSSConfig 默认 XSS 配置
func DefaultXSSConfig() XSSConfig {
	return XSSConfig{
		SkipPaths: []string{
			"/api/health",
			"/api/realtime",
		},
		SkipPrefixes: []string{
			"/api/files",
			"/api/backups",
			"/shop",      // 商城静态页面
			"/assets",    // 静态资源
		},
		SkipQueryKeys: []string{
			"filter", // filter 参数用于数据库查询，不需要 HTML 转义
		},
		StripTags:  false,
		EscapeHTML: true,
	}
}

// XSSMiddleware XSS 防护中间件
func XSSMiddleware(config ...XSSConfig) gin.HandlerFunc {
	cfg := DefaultXSSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 检查是否跳过
		path := c.Request.URL.Path
		for _, skipPath := range cfg.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}
		for _, prefix := range cfg.SkipPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		// 设置安全响应头
		setSecurityHeaders(c)

		// 对请求参数进行 XSS 过滤
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// JSON 请求体处理
				sanitizeJSONBody(c, cfg)
			} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
				// 表单数据处理
				sanitizeFormData(c, cfg)
			}
		}

		// 查询参数过滤
		sanitizeQueryParams(c, cfg)

		c.Next()
	}
}

// setSecurityHeaders 设置安全响应头
func setSecurityHeaders(c *gin.Context) {
	// 防止 MIME 类型嗅探
	c.Header("X-Content-Type-Options", "nosniff")
	// 防止点击劫持
	c.Header("X-Frame-Options", "DENY")
	// XSS 保护
	c.Header("X-XSS-Protection", "1; mode=block")
	// 内容安全策略（简化版）
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:;")
	// 引用策略
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	// 权限策略
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// sanitizeJSONBody 过滤 JSON 请求体
// 注意：JSON 请求体不应进行 HTML 转义，否则会破坏 JSON 格式
// JSON 字符串值已经有自己的转义机制
func sanitizeJSONBody(c *gin.Context, cfg XSSConfig) {
	if c.Request.Body == nil {
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil {
		return
	}

	// 对 JSON 请求体，只移除危险内容，不进行 HTML 转义
	// 因为 HTML 转义会破坏 JSON 格式（如 & -> &amp;）
	sanitized := string(bodyBytes)

	// 只移除危险的 JavaScript 事件和协议，不转义 HTML
	sanitized = removeDangerousEvents(sanitized)
	sanitized = removeDangerousProtocols(sanitized)

	// 替换请求体
	c.Request.Body = io.NopCloser(bytes.NewBufferString(sanitized))
}

// sanitizeFormData 过滤表单数据
func sanitizeFormData(c *gin.Context, cfg XSSConfig) {
	if err := c.Request.ParseForm(); err != nil {
		return
	}

	for key, values := range c.Request.PostForm {
		for i, value := range values {
			c.Request.PostForm[key][i] = xssSanitize(value, cfg)
		}
	}
}

// sanitizeQueryParams 过滤查询参数
func sanitizeQueryParams(c *gin.Context, cfg XSSConfig) {
	query := c.Request.URL.Query()
	for key, values := range query {
		// 检查是否跳过该参数
		skip := false
		for _, skipKey := range cfg.SkipQueryKeys {
			if key == skipKey {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for i, value := range values {
			query[key][i] = xssSanitize(value, cfg)
		}
	}
	c.Request.URL.RawQuery = query.Encode()
}

// xssSanitize XSS 过滤核心函数
func xssSanitize(input string, cfg XSSConfig) string {
	if input == "" {
		return input
	}

	// URL 解码
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		decoded = input
	}

	// 移除危险的 HTML 标签
	if cfg.StripTags {
		decoded = stripTags(decoded)
	}

	// 转义 HTML 字符
	if cfg.EscapeHTML {
		decoded = html.EscapeString(decoded)
	}

	// 移除危险的 JavaScript 事件
	decoded = removeDangerousEvents(decoded)

	// 移除危险的协议
	decoded = removeDangerousProtocols(decoded)

	return decoded
}

// stripTags 移除 HTML 标签
func stripTags(input string) string {
	// 移除所有 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// removeDangerousEvents 移除危险的 JavaScript 事件
func removeDangerousEvents(input string) string {
	// 常见的危险事件属性
	events := []string{
		"onclick", "ondblclick", "onmousedown", "onmouseup", "onmouseover",
		"onmousemove", "onmouseout", "onmouseenter", "onmouseleave",
		"onkeydown", "onkeypress", "onkeyup",
		"onfocus", "onblur", "onchange", "onsubmit", "onreset",
		"onload", "onunload", "onerror", "onabort",
		"ondrag", "ondragend", "ondragenter", "ondragleave", "ondragover", "ondragstart", "ondrop",
		"onscroll", "onresize", "oncontextmenu",
		"onanimationstart", "onanimationend", "onanimationiteration",
		"ontransitionend",
	}

	result := input
	for _, event := range events {
		// 移除事件属性（不区分大小写）
		re := regexp.MustCompile(`(?i)\s*` + event + `\s*=\s*["'][^"']*["']`)
		result = re.ReplaceAllString(result, "")
	}

	return result
}

// removeDangerousProtocols 移除危险的协议
func removeDangerousProtocols(input string) string {
	// 危险协议
	protocols := []string{
		"javascript:", "vbscript:", "data:text/html", "data:application",
	}

	result := input
	for _, protocol := range protocols {
		// 移除危险协议（不区分大小写）
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(protocol))
		result = re.ReplaceAllString(result, "")
	}

	return result
}

// XSSSanitize 公开的 XSS 过滤函数
func XSSSanitize(input string) string {
	return xssSanitize(input, DefaultXSSConfig())
}

// XSSSanitizeWithConfig 带配置的 XSS 过滤函数
func XSSSanitizeWithConfig(input string, cfg XSSConfig) string {
	return xssSanitize(input, cfg)
}
