package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/pkg/auth"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware 可选的用户认证中间件
// 如果提供了Authorization header，则验证并设置用户信息；否则继续执行
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证信息，继续执行（可能是公开操作）
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 格式错误，继续执行（不强制认证）
			c.Next()
			return
		}

		token := parts[1]
		if token == "" {
			c.Next()
			return
		}

		// 先尝试解析为UserClaims（Auth集合用户token）
		claims, err := auth.ValidateUserToken(token)
		if err != nil {
			// 如果不是UserClaims，尝试解析为通用Claims
			claims2, err2 := auth.ValidateToken(token)
			if err2 != nil {
				// token无效，继续执行（不强制认证）
				c.Next()
				return
			}
			// 设置通用用户信息
			c.Set("user_id", claims2.UserID)
			c.Set("user_email", claims2.Email)
			c.Set("user_role", claims2.Role)
			c.Next()
			return
		}

		// 设置Auth集合用户信息
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_collection", claims.Collection)
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, "Token is required")
			c.Abort()
			return
		}

		claims, err := auth.ValidateAdminToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_email", claims.Email)
		c.Next()
	}
}
