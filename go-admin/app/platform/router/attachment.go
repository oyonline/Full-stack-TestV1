package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/platform/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerAttachmentRouter)
}

func registerAttachmentRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Attachment{}
	r := v1.Group("/platform/attachments").
		Use(authMiddleware.MiddlewareFunc()).
		Use(middleware.AuthCheckRole()).
		Use(actions.PermissionAction())
	{
		r.GET("", api.GetPage)
		r.POST("/upload", api.Upload)
		r.GET("/:id/download", api.Download)
		r.DELETE("/:id", api.Delete)
	}
}
