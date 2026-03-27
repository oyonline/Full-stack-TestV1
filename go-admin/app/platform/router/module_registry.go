package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/platform/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerModuleRegistryRouter)
}

func registerModuleRegistryRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.ModuleRegistry{}
	r := v1.Group("/platform/modules").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole()).
		Use(actions.PermissionAction())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("", api.Update)
		r.DELETE("/:id", api.Delete)
	}
}
