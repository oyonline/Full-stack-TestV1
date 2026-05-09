package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSpuRouter)
}

// registerSpuRouter SPU 路由。
//
// PermissionAction 中间件注入 dataScope，service 层通过 actions.GetPermissionFromContext 取出后做范围过滤。
func registerSpuRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Spu{}
	r := v1.Group("/spu").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole()).
		Use(actions.PermissionAction())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("", api.Delete)
		r.POST("/:id/submit", api.Submit)
		r.POST("/:id/offline", api.GoOffline)
		r.POST("/:id/online", api.GoOnline)
	}
}
