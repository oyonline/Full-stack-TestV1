package service

import (
	"go-admin/app/admin/models"
	"go-admin/common/baseservice"
)

// SysLoginLog 使用 BaseService[models.SysLoginLog] 提供 GetPage/Get/Remove。
// （历史代码只暴露了这三件套；BaseService 多出来的 Insert/Update 不会被路由到，因此无副作用。）
type SysLoginLog struct {
	baseservice.BaseService[models.SysLoginLog]
}
