//go:build embed
// +build embed

package embed

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var staticFS embed.FS

func init() {
	// 标记已嵌入文件
	setEmbedded()
}

var embedded = false

func setEmbedded() {
	embedded = true
}

func hasEmbeddedFiles() bool {
	return embedded
}

func getEmbeddedFS() http.FileSystem {
	fsys, _ := fs.Sub(staticFS, "dist")
	return http.FS(fsys)
}

// serveEmbedded 使用嵌入的文件服务
func serveEmbedded(r *gin.Engine) {
	distFS, _ := fs.Sub(staticFS, "dist")

	// 静态资源服务
	r.StaticFS("/assets", http.FS(distFS))

	// SPA 路由 fallback
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

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
			// 尝试从嵌入的文件系统读取
			file, err := distFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				defer file.Close()
				stat, _ := file.Stat()
				http.ServeContent(c.Writer, c.Request, path, stat.ModTime(), file.(io.ReadSeeker))
				return
			}
		}

		// 其他路由返回 index.html (SPA)
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index.html not found")
			return
		}
		defer indexFile.Close()

		stat, _ := indexFile.Stat()
		http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})
}

// readEmbeddedFile 读取嵌入的文件
func readEmbeddedFile(filename string) ([]byte, error) {
	return fs.ReadFile(staticFS, "dist/"+filename)
}

// getEmbeddedIndexHTML 获取嵌入的 index.html
func getEmbeddedIndexHTML() (io.ReadCloser, error) {
	return staticFS.Open("dist/index.html")
}
