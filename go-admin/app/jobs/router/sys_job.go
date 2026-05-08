package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/jobs/apis"
	models2 "go-admin/app/jobs/models"
	dto2 "go-admin/app/jobs/service/dto"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSysJobRouter)
}

// 需认证的路由代码
func registerSysJobRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {

	r := v1.Group("/sysjob").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	{
		// 平台底座：作业管理由超管操作，不挂 PermissionAction。
		// IndexAction/ViewAction 内部仍调 GetPermissionFromContext + Permission()，但中间件未挂时
		// p 为空 DataScope，Permission() 走 default 分支返回原 db，无过滤效果。
		// 老框架内部 scope 解除待 phase3 评估（涉及通用 helper 改动）。
		sysJob := &models2.SysJob{}
		r.GET("", actions.IndexAction(sysJob, new(dto2.SysJobSearch), func() interface{} {
			list := make([]models2.SysJob, 0)
			return &list
		}))
		r.GET("/:id", actions.ViewAction(new(dto2.SysJobById), func() interface{} {
			return &dto2.SysJobItem{}
		}))
		r.POST("", actions.CreateAction(new(dto2.SysJobControl)))
		r.PUT("", actions.UpdateAction(new(dto2.SysJobControl)))
		r.DELETE("", actions.DeleteAction(new(dto2.SysJobById)))
	}
	sysJob := apis.SysJob{}

	v1.GET("/job/remove/:id", sysJob.RemoveJobForService)
	v1.GET("/job/start/:id", sysJob.StartJobForService)
}
