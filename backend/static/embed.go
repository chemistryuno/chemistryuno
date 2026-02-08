package static

import (
	"embed"
	"io/fs"
)

//go:embed dist
var DistFS embed.FS

// GetDistFS 返回前端构建文件的文件系统
func GetDistFS() (fs.FS, error) {
	return fs.Sub(DistFS, "dist")
}
