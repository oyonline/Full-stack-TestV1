package version

import (
	"runtime"
	"strconv"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1775300000000RemoveDictDataPage)
}

// _1775300000000RemoveDictDataPage 收口"字典数据"作为独立菜单页的历史方案：
//
//   - menu_id=59 (SysDictDataManage, component='/admin/sys-dict-data/index') 已无对应前端页，
//     是字典菜单'假已完成'的根因；删除后侧栏只保留 menu_id=543 '字典类型'。
//   - menu_id=240 ('查询数据' F-按钮，permission=admin:sysDictData:query) 与 59 配套，一并删除。
//   - menu_id=241/242/243 ('新增/修改/删除数据') 仍保留为字典数据增改删按钮，但要从 59 的子级
//     迁移到 543 ('字典类型') 之下，对齐当前真实左侧导航。
//   - sys_role_menu / sys_menu_api_rule 中以 59、240 为左侧的关联记录一并清理。
//
// 设计说明：
//   - 仅按 menu_id 命中固定值（59、240、241、242、243），不依赖 path/permission 前缀，避免误伤。
//   - 不动 sys_dict 后端路由（/api/v1/dict/data 仍提供，给字典类型页内嵌 tab 复用）。
//   - 不动 sys_casbin_rule（V1 是 API path，不绑定 menu_id；菜单重排不影响策略命中）。
//   - 全部 DELETE / UPDATE 用条件命中，重跑无副作用。
func _1775300000000RemoveDictDataPage(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		removedMenuIDs := []int{59, 240}
		movedMenuIDs := []int{241, 242, 243}
		const newParentID = 543
		const newParentPaths = "/0/2/58/543"

		if err := tx.Exec(
			"DELETE FROM sys_role_menu WHERE menu_id IN ?",
			removedMenuIDs,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			"DELETE FROM sys_menu_api_rule WHERE sys_menu_menu_id IN ?",
			removedMenuIDs,
		).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			"DELETE FROM sys_menu WHERE menu_id IN ?",
			removedMenuIDs,
		).Error; err != nil {
			return err
		}

		for _, id := range movedMenuIDs {
			newPaths := newParentPaths + "/" + strconv.Itoa(id)
			if err := tx.Exec(
				"UPDATE sys_menu SET parent_id = ?, paths = ? WHERE menu_id = ?",
				newParentID, newPaths, id,
			).Error; err != nil {
				return err
			}
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
