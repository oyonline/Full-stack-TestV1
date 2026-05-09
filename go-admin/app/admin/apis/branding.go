package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-admin/internal/branding"
)

// GetEmailPreview renders an email template with an inline base64 logo and returns HTML.
// Used by e2e tests to verify T7 email logo embedding without sending a real email.
// @Summary 渲染邮件模板预览 HTML
// @Tags branding
// @Param appName query string false "系统名称，默认 System"
// @Param title query string false "邮件标题"
// @Param bg query string false "品牌底色 #RRGGBB，默认 #1d4ed8"
// @Produce text/html
// @Success 200 {string} string "HTML content"
// @Router /api/v1/branding/email-preview [get]
func (Branding) GetEmailPreview(c *gin.Context) {
	appName := c.DefaultQuery("appName", "System")
	title := c.DefaultQuery("title", "测试邮件")
	body := c.DefaultQuery("body", "<p>这是一封测试邮件。</p>")
	footer := c.DefaultQuery("footer", "")
	bgHex := c.DefaultQuery("bg", "#1d4ed8")

	html, err := branding.RenderEmailTemplate(appName, title, body, footer, bgHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

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
