package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-admin/internal/branding"
)

type Branding struct{}

// GetDefaultLogoPNG renders a fallback logo PNG and returns it.
// @Summary 获取默认 logo PNG
// @Description 返回首字 fallback logo，image/png 格式
// @Tags branding
// @Param text query string false "单字符文本，默认 S"
// @Param bg query string false "背景色 #RRGGBB，默认 #1d4ed8"
// @Param size query int false "尺寸（16/32/64/96/128/256），默认 64"
// @Param v query string false "版本信号，用于 ETag"
// @Produce image/png
// @Success 200 {file} binary
// @Router /api/v1/branding/default-logo.png [get]
func (Branding) GetDefaultLogoPNG(c *gin.Context) {
	text := c.DefaultQuery("text", "S")
	bgHex := c.DefaultQuery("bg", "#1d4ed8")
	sizeStr := c.DefaultQuery("size", "64")
	v := c.Query("v")

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size parameter"})
		return
	}

	etag := branding.ETagFor(text, bgHex, size, v)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	data, err := branding.RenderDefaultLogoPNG(text, bgHex, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("ETag", etag)
	c.Data(http.StatusOK, "image/png", data)
}
