package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

const (
	csrfTokenLength = 32
	csrfHeader      = "X-CSRF-Token"
	csrfCookieName  = "csrf_token"
)

// CSRFConfig CSRF 中间件配置
type CSRFConfig struct {
	SkipPaths    []string // 跳过 CSRF 验证的路径
	SkipPrefixes []string // 跳过 CSRF 验证的路径前缀
	SecureCookie bool     // 是否设置 Secure 标志
	SameSite     string   // SameSite 属性
}

// DefaultCSRFConfig 默认 CSRF 配置
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		SkipPaths: []string{
			"/api/health",
			"/api/admins/auth-with-password",
			"/api/admins/auth-refresh",
		},
		SkipPrefixes: []string{
			"/api/collections/",
			"/api/files",
			"/api/realtime",
			"/api/backups",
		},
		SecureCookie: false,
		SameSite:     "Lax",
	}
}

// CSRFMiddleware CSRF 保护中间件
func CSRFMiddleware(config ...CSRFConfig) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 只对修改操作进行 CSRF 验证
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			// 生成并设置 CSRF token
			token := generateCSRFToken()
			c.Set("csrf_token", token)

			// 设置 cookie
			c.SetCookie(
				csrfCookieName,
				token,
				3600*24, // 24小时
				"/",
				"",
				cfg.SecureCookie,
				true, // HttpOnly
			)

			// 设置响应头
			c.Header(csrfHeader, token)
			c.Next()
			return
		}

		// 检查是否跳过验证
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

		// 从请求头获取 CSRF token
		tokenFromHeader := c.GetHeader(csrfHeader)
		if tokenFromHeader == "" {
			// 尝试从表单获取
			tokenFromHeader = c.PostForm("_csrf")
		}

		// 从 cookie 获取 CSRF token
		tokenFromCookie, err := c.Cookie(csrfCookieName)
		if err != nil || tokenFromCookie == "" {
			response.Error(c, http.StatusForbidden, "CSRF token not found in cookie")
			c.Abort()
			return
		}

		// 验证 token
		if tokenFromHeader != tokenFromCookie {
			response.Error(c, http.StatusForbidden, "CSRF token mismatch")
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateCSRFToken 生成 CSRF token
func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳作为后备
		return base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", csrfTokenLength)))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// GetCSRFToken 从上下文获取 CSRF token
func GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		return token.(string)
	}
	return ""
}
