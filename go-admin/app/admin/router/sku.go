package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSkuRouter)
}

// registerSkuRouter SKU 路由（只读）。
// 写动作（add/edit/remove）通过 SPU 编辑页的子表完成，此处不暴露。
func registerSkuRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Sku{}
	r := v1.Group("/sku").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole()).
		Use(actions.PermissionAction())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
	}
}
