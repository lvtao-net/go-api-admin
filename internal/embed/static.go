//go:build !embed
// +build !embed

package embed

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// StaticFS 返回静态文件系统（根据是否嵌入选择不同的实现）
func StaticFS(useEmbed bool) http.FileSystem {
	if useEmbed {
		return embeddedStaticFS()
	}
	return http.Dir("./web/dist")
}

// embeddedStaticFS 返回嵌入的静态文件系统
// 注意：嵌入功能在编译时通过 -tags embed 构建
func embeddedStaticFS() http.FileSystem {
	// 检查是否有嵌入的文件系统
	if hasEmbeddedFiles() {
		return getEmbeddedFS()
	}
	// 回退到外部文件
	return http.Dir("./web/dist")
}

// hasEmbeddedFiles 检查是否有嵌入的文件
func hasEmbeddedFiles() bool {
	// 检查是否通过 embed tag 构建
	return false // 默认返回 false，使用外部文件
}

// getEmbeddedFS 获取嵌入的文件系统（由 embed 构建标签版本提供）
func getEmbeddedFS() http.FileSystem {
	return http.Dir("./web/dist") // 默认实现
}

// ServeStatic 配置静态文件服务和 SPA 路由
// useEmbed: true 使用嵌入的静态文件，false 使用外部 web/dist 目录
func ServeStatic(r *gin.Engine, useEmbed bool) {
	var staticDir string
	if useEmbed && hasEmbeddedFiles() {
		staticDir = "embedded"
	} else {
		staticDir = "./web/dist"
	}

	// 使用外部文件系统（开发模式或无嵌入时）
	if staticDir == "./web/dist" {
		// 检查 web/dist 目录是否存在
		if _, err := os.Stat("./web/dist"); os.IsNotExist(err) {
			// 如果不存在，不配置静态文件服务
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
	} else {
		// 嵌入模式（使用内存文件系统）
		serveEmbedded(r)
	}
}

// serveEmbedded 使用嵌入的文件服务
func serveEmbedded(r *gin.Engine) {
	// 这个函数由 embed 构建标签版本实现
	// 默认回退到外部文件
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Static files not embedded")
	})
}

// IsEmbedded 检查是否成功嵌入了静态文件
func IsEmbedded() bool {
	return hasEmbeddedFiles()
}

// ReadFile 读取文件内容（支持嵌入模式）
func ReadFile(filename string) ([]byte, error) {
	if hasEmbeddedFiles() {
		return readEmbeddedFile(filename)
	}
	return os.ReadFile("./web/dist/" + filename)
}

// readEmbeddedFile 读取嵌入的文件（由 embed 构建标签版本提供）
func readEmbeddedFile(filename string) ([]byte, error) {
	return nil, fs.ErrNotExist
}

// GetIndexHTML 获取 index.html 内容
func GetIndexHTML() (io.ReadCloser, error) {
	if hasEmbeddedFiles() {
		return getEmbeddedIndexHTML()
	}
	return os.Open("./web/dist/index.html")
}

// getEmbeddedIndexHTML 获取嵌入的 index.html（由 embed 构建标签版本提供）
func getEmbeddedIndexHTML() (io.ReadCloser, error) {
	return nil, fs.ErrNotExist
}
