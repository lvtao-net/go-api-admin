//go:build !embed
// +build !embed

package embed

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// hasEmbeddedFiles 检查是否有嵌入的文件
func hasEmbeddedFiles() bool {
	return false
}

// getEmbeddedFS 获取嵌入的文件系统
func getEmbeddedFS() http.FileSystem {
	return http.Dir("./web/dist")
}

// ServeStatic 配置静态文件服务和 SPA 路由
func ServeStatic(r *gin.Engine, useEmbed bool) {
	// 检查 web/dist 目录是否存在
	if _, err := os.Stat("./web/dist"); os.IsNotExist(err) {
		return
	}

	// 静态资源服务
	r.Static("/assets", "./web/dist/assets")

	// SPA 路由 fallback
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是 API 路由，返回 404 JSON 响应
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "API endpoint not found",
			})
			return
		}

		// 检查是否是静态资源
		staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
		isStatic := false
		for _, ext := range staticExts {
			if strings.HasSuffix(path, ext) {
				isStatic = true
				break
			}
		}

		if isStatic {
			c.File("./web/dist" + path)
			return
		}

		// 其他路由返回 index.html (SPA)
		c.File("./web/dist/index.html")
	})
}

// readEmbeddedFile 读取嵌入的文件
func readEmbeddedFile(filename string) ([]byte, error) {
	return nil, os.ErrNotExist
}

// getEmbeddedIndexHTML 获取嵌入的 index.html
func getEmbeddedIndexHTML() (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}
