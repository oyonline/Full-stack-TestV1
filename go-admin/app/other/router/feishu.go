package router

import (
	"go-admin/app/other/apis"

	"github.com/gin-gonic/gin"
)

func init() {
	routerNoCheckRole = append(routerNoCheckRole, registerFeishuRouter)
}

// 需认证的路由代码
func registerFeishuRouter(v1 *gin.RouterGroup) {
	api := apis.FeishuCallback{}
	r := v1.Group("/feishu")
	{
		r.POST("/callback", api.Callback)
		r.POST("/orgList", api.OrgList)
		r.POST("/departtment", api.DepartmentList)
		r.POST("/platform", api.PlatformList)
		r.POST("/feeType", api.FeeCode)
		r.GET("/subscript", api.Subscript)
	}
}
