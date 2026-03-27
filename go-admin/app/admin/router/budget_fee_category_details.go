package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerBudgetFeeCategoryDetailsRouter)
}

// registerBudgetFeeCategoryDetailsRouter
func registerBudgetFeeCategoryDetailsRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.BudgetFeeCategoryDetails{}
	r := v1.Group("/budget-fee-category-details").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	{
		r.GET("/list", actions.PermissionAction(), api.GetPage)
		r.GET("/get/:id", actions.PermissionAction(), api.Get)
		r.POST("/add", actions.PermissionAction(), api.Insert)
		r.PUT("/edit/:id", actions.PermissionAction(), api.Update)
		r.DELETE("/remove", actions.PermissionAction(), api.Delete)

		// 新增导入导出功能路由
		r.POST("/import", actions.PermissionAction(), api.Import)
		r.GET("/export", actions.PermissionAction(), api.Export)
		r.GET("/template", actions.PermissionAction(), api.DownloadTemplate)
	}
}
