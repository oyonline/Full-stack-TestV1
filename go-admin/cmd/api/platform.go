package api

import "go-admin/app/platform/router"

func init() {
	// 注册平台能力路由
	AppRouters = append(AppRouters, router.InitRouter)
}
