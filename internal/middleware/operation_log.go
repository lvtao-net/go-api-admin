package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/service"
)

// responseWriter 用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录写操作 (POST, PATCH, DELETE)
		method := c.Request.Method
		if method != "POST" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// 排除不需要记录的路径
		path := c.Request.URL.Path
		if strings.Contains(path, "/auth-refresh") ||
			strings.Contains(path, "/auth-with-password") ||
			strings.Contains(path, "/admins/logs") ||
			strings.Contains(path, "/api/health") {
			c.Next()
			return
		}

		// 捕获请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		// 记录开始时间
		start := time.Now()

		c.Next()

		// 获取用户信息
		var userID uint64
		var userEmail, userType string
		if adminID, exists := c.Get("admin_id"); exists {
			switch v := adminID.(type) {
			case uint64:
				userID = v
			case float64:
				userID = uint64(v)
			case string:
				// 兼容旧的string类型
			}
			userType = "admin"
			if adminEmail, exists := c.Get("admin_email"); exists {
				userEmail = adminEmail.(string)
			}
		} else if uid, exists := c.Get("user_id"); exists {
			switch v := uid.(type) {
			case uint64:
				userID = v
			case float64:
				userID = uint64(v)
			case string:
				// 兼容旧的string类型
			}
			userType = "user"
			if uemail, exists := c.Get("user_email"); exists {
				userEmail = uemail.(string)
			}
		}

		// 确定操作类型
		action := getAction(method, path)
		if action == "" {
			return
		}

		// 提取集合名称
		collection := extractCollection(path)

		// 创建日志记录
		log := model.OperationLog{
			Collection: collection,
			Action:     action,
			UserID:     userID,
			UserEmail:  userEmail,
			UserType:   userType,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Request:    string(requestBody),
			Response:   writer.body.String(),
			Status:     c.Writer.Status(),
		}
		log.Created = start
		log.Updated = start

		// 异步记录日志
		logService := service.NewLogService()
		logService.Log(log)
	}
}

// getAction 根据请求方法和路径获取操作类型
func getAction(method, path string) string {
	switch method {
	case "POST":
		if strings.Contains(path, "/auth-with-password") || strings.Contains(path, "/login") {
			return "login"
		}
		if strings.Contains(path, "/register") {
			return "register"
		}
		return "create"
	case "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

// extractCollection 从路径中提取集合名称
func extractCollection(path string) string {
	// /api/collections/{collection}/records
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "collections" && i+1 < len(parts) {
			collection := parts[i+1]
			// 排除一些特殊路径
			if collection != "records" && !strings.Contains(collection, "{") {
				return collection
			}
		}
	}
	return ""
}

// PrettyJSON 格式化JSON
func PrettyJSON(data string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return data
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return data
	}
	return string(pretty)
}
