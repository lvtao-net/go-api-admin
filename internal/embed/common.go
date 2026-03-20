package embed

import (
	"io"
	"net/http"
	"os"
)

// StaticFS 返回静态文件系统（根据是否嵌入选择不同的实现）
func StaticFS(useEmbed bool) http.FileSystem {
	if useEmbed && hasEmbeddedFiles() {
		return getEmbeddedFS()
	}
	return http.Dir("./web/dist")
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

// GetIndexHTML 获取 index.html 内容
func GetIndexHTML() (io.ReadCloser, error) {
	if hasEmbeddedFiles() {
		return getEmbeddedIndexHTML()
	}
	return os.Open("./web/dist/index.html")
}
