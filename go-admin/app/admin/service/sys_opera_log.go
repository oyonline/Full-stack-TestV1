package service

import (
	"go-admin/app/admin/models"
	"go-admin/common/baseservice"
)

// SysOperaLog 嵌入 BaseService 拿到 GetPage/Get/Remove 默认实现。
// Insert 与 BaseService 的签名不同（直接吃 *models 而非 DTO），保留自定义实现以兼容调用方。
type SysOperaLog struct {
	baseservice.BaseService[models.SysOperaLog]
}

// Insert 创建 SysOperaLog 对象。
// 中间件 SetDBOperLog 会把 map[string]interface{} 转 model 后调用此方法，签名与 BaseService.Insert 不同。
func (e *SysOperaLog) Insert(model *models.SysOperaLog) error {
	if err := e.Orm.Create(model).Error; err != nil {
		e.Log.Errorf("Service InsertSysOperaLog error:%s", err.Error())
		return err
	}
	return nil
}
