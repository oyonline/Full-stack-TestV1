package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/common/middleware"

	"go-admin/app/admin/apis"
	"go-admin/common/actions"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerAllocationRuleSettingsRouter)
}

// registerAllocationRuleSettingsRouter
func registerAllocationRuleSettingsRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.AllocationRuleSettings{}
	r := v1.Group("/allocation-rule-settings").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	{
		r.GET("/list", actions.PermissionAction(), api.GetPage)
		r.GET("/get/:id", actions.PermissionAction(), api.Get)
		r.POST("/add", api.Insert)
		r.PUT("/edit/:id", actions.PermissionAction(), api.Update)
		r.DELETE("/remove", api.Delete)

		// 新增导入导出功能路由
		r.POST("/import", actions.PermissionAction(), api.Import)
		r.GET("/export", actions.PermissionAction(), api.Export)
		r.GET("/template", actions.PermissionAction(), api.DownloadTemplate)
	}
}
