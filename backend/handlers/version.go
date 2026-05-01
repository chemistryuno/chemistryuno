package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetVersion 获取应用版本信息
func GetVersion(c *gin.Context) {
	version := os.Getenv("APP_VERSION")
	versionName := os.Getenv("APP_VERSION_NAME")

	// 如果环境变量未设置，使用默认值
	if version == "" {
		version = "1.2.1"
	}
	if versionName == "" {
		versionName = "Mendeleef"
	}

	c.JSON(http.StatusOK, gin.H{
		"version":     version,
		"versionName": versionName,
		"fullVersion": "V" + version + " " + versionName,
	})
}
