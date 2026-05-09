package router

import (
	"go-admin/app/admin/apis"

	"github.com/gin-gonic/gin"
)

func init() {
	routerNoCheckRole = append(routerNoCheckRole, registerBrandingRouter)
}

func registerBrandingRouter(v1 *gin.RouterGroup) {
	api := apis.Branding{}
	r := v1.Group("/branding")
	{
		r.GET("/default-logo.png", api.GetDefaultLogoPNG)
	}
}
