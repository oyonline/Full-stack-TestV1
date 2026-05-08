package service

import (
	"errors"

	log "github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/gorm"

	platformModels "go-admin/app/platform/models"
)

// EnsureModuleEnabled 校验业务模块是否注册并启用。
//
// 严格语义：表中存在一条 module_key=<key> AND status=2 的记录 → 放行。
//
// 兜底语义（module_registry 整张表里没有任何 status=2 的记录时）：
// 视为"模块管控尚未启用"，放行并 log.Warn 提示。表非空但不命中传入 key 的情况下仍按原逻辑拒绝。
//
// 判定基准是 "status=2 的记录数" 而非"行数"——只要全表没有任何启用的模块，
// 就等价于"空表"语义；这样避免一行 status=1 的禁用记录把整张表卡死。
//
// 该兜底是 phase2 之前的临时降级，待 module_registry 正式接入后应移除。
func EnsureModuleEnabled(orm *gorm.DB, moduleKey string) error {
	var m platformModels.ModuleRegistry
	err := orm.Where("module_key = ? AND status = ?", moduleKey, "2").First(&m).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var enabledCount int64
	if err := orm.Model(&platformModels.ModuleRegistry{}).Where("status = ?", "2").Count(&enabledCount).Error; err != nil {
		return err
	}
	if enabledCount == 0 {
		log.Warnf("[module_gate] module_registry 为空，降级放行 moduleKey=%s（建议 phase2 接入模块注册时移除此兜底）", moduleKey)
		return nil
	}
	return errors.New("模块未注册或未启用")
}
