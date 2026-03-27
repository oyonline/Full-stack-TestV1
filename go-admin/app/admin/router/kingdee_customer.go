package router

import (
	"go-admin/app/admin/apis"
	"go-admin/common/middleware"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerKingdeeCustomerRouter)
}

// 需认证的路由代码
func registerKingdeeCustomerRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.KingdeeCustomer{}
	r := v1.Group("/kingdee-customer").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("", api.Delete)
		r.GET("/template", api.DownloadTemplate)
		r.POST("/import", api.Import)
		r.GET("/export", api.Export)

		r.GET("/pull", api.PullKingdeeCustomers)
		r.POST("/group", api.PullKingdeeCustomerGroups)
		r.GET("/group", api.GetKingdeeCustomerGroups)
		r.GET("/organize", api.PullKingdeeOrganizeInfos)
	}
}
