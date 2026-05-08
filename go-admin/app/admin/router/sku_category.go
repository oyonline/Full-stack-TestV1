package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/admin/apis"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSkuCategoryRouter)
}

func registerSkuCategoryRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.SkuCategory{}
	r := v1.Group("/sku-category").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole())
	{
		r.GET("", api.GetTree)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("", api.Delete)
	}
}
