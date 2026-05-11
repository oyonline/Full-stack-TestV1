package service

import (
	"errors"

	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
)

// EnsureModuleEnabled 校验业务模块是否注册并启用。
//
// 严格语义：表中必须存在一条 module_key=<key> AND status=2 的记录，否则返回错误。
// 不再提供兜底放行；调用方应确保 module_registry 已正确注入对应模块。
func EnsureModuleEnabled(orm *gorm.DB, moduleKey string) error {
	var m platformModels.ModuleRegistry
	err := orm.Where("module_key = ? AND status = ?", moduleKey, "2").First(&m).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("模块未注册或未启用")
	}
	return err
}
