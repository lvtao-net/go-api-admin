package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/config"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	requests map[string]*clientLimit
	mu       sync.RWMutex
	rate     int           // 每分钟请求数
	burst    int           // 突发容量
	window   time.Duration // 时间窗口
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter() *RateLimiter {
	cfg := config.GetRateLimitSettings()
	rate := 60
	burst := 10
	if r, ok := cfg["requestsPerMinute"].(int); ok {
		rate = r
	}
	if b, ok := cfg["burst"].(int); ok {
		burst = b
	}

	rl := &RateLimiter{
		requests: make(map[string]*clientLimit),
		rate:     rate,
		burst:    burst,
		window:   time.Minute,
	}

	// 启动清理过期记录的goroutine
	go rl.cleanup()

	return rl
}

var (
	rateLimiter *RateLimiter
	once        sync.Once
)

// GetRateLimiter 获取速率限制器单例
func GetRateLimiter() *RateLimiter {
	once.Do(func() {
		rateLimiter = NewRateLimiter()
	})
	return rateLimiter
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	client, exists := r.requests[key]

	if !exists || now.After(client.resetTime) {
		// 新客户端或窗口已过期
		r.requests[key] = &clientLimit{
			count:     1,
			resetTime: now.Add(r.window),
		}
		return true
	}

	if client.count < r.rate+r.burst {
		client.count++
		return true
	}

	return false
}

// cleanup 清理过期的记录
func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for key, client := range r.requests {
			if now.After(client.resetTime) {
				delete(r.requests, key)
			}
		}
		r.mu.Unlock()
	}
}

// RateLimitMiddleware 速率限制中间件（基于IP）
func RateLimitMiddleware() gin.HandlerFunc {
	rl := GetRateLimiter()

	return func(c *gin.Context) {
		// 获取客户端标识（IP或用户ID）
		key := c.ClientIP()

		if !rl.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, "Too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimitMiddleware 速率限制中间件（基于用户ID）
func UserRateLimitMiddleware() gin.HandlerFunc {
	rl := GetRateLimiter()

	return func(c *gin.Context) {
		// 尝试获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，使用IP
			userID = c.ClientIP()
		}

		key := "user:" + userID.(string)

		if !rl.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
