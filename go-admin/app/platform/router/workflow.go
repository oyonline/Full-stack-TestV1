package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/platform/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerWorkflowRouter)
}

func registerWorkflowRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.Workflow{}
	r := v1.Group("/platform/workflow").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole()).Use(actions.PermissionAction())
	{
		r.GET("/definitions", api.GetDefinitionPage)
		r.GET("/definitions/:id", api.GetDefinition)
		r.POST("/definitions", api.InsertDefinition)
		r.PUT("/definitions", api.UpdateDefinition)
		r.DELETE("/definitions/:id", api.DeleteDefinition)
		r.GET("/definitions/:id/nodes", api.GetDefinitionNodes)
		r.PUT("/definitions/:id/nodes", api.SaveDefinitionNodes)

		r.POST("/instances/start", api.StartInstance)
		r.GET("/instances/started", api.GetStartedInstancePage)
		r.GET("/instances/:id", api.GetInstance)
		r.POST("/instances/:id/withdraw", api.WithdrawInstance)

		r.GET("/tasks/todo", api.GetTodoTaskPage)
		r.POST("/tasks/:id/approve", api.ApproveTask)
		r.POST("/tasks/:id/reject", api.RejectTask)
	}
}
