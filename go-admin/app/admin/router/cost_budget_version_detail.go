package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerCostBudgetVersionDetailRouter)
}

// registerCostBudgetVersionDetailRouter
func registerCostBudgetVersionDetailRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.CostBudgetVersionDetail{}
	r := v1.Group("/cost-budget-version-detail").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
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
