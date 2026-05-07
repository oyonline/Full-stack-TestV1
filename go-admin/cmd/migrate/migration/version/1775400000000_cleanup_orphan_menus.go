package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1775400000000CleanupOrphanMenus)
}

// _1775400000000CleanupOrphanMenus 清理 sys_menu 中真实 orphan 菜单。
//
// 通用方法论（沿用 my-c34 / 1775300000000_remove_dict_data_page）：
//
//	对每个待删菜单：
//	  - 若有子节点，先把子节点 reparent 到祖父级再 DELETE 父节点；
//	  - 若是叶子节点，直接 DELETE 即可。
//	无论哪种情况，同步清理 sys_role_menu / sys_menu_api_rule 中以该 menu_id
//	为左侧的关联记录。
//
// 本次应用：原 spec 列出 7 个 menu_id（262/61/211/460/471/269/537）声称是
// orphan，经 mayor + fury 独立比对当前 db.sql 与前端
// vue-vben-admin/apps/web-antd/src/views/ 实际视图后（witness 升级
// hq-wisp-b9m），结论是只有 menu_id=471 (JobLog '/schedule/log') 是真叶子
// orphan：
//
//   - 262 EditTable -> /dev-tools/gen/edit (edit.vue 存在)
//   - 61 Swagger    -> /admin/sys-api/index (sys-api/index.vue 存在)
//   - 211 Log       -> RouteView 父级，子 212/216 是真实日志页
//   - 460 ScheduleManage -> /admin/sys-job/index (sys-job/index.vue 存在)
//   - 471 JobLog    -> /schedule/log，views/schedule/ 整个目录不存在 ← 删
//   - 269 ServerMonitor  -> /admin/sys-server-monitor/index (存在)
//   - 537 SysTools  -> Layout 父级，子 269 + 680 WorkflowCenter 全是真页
//
// 471 无子节点，按方法论分支二（leaf）处理，仅 DELETE，不需要 reparent。
// 父级 459 ('定时任务' Schedule Layout) 与兄弟 460 (ScheduleManage) 全部保留。
func _1775400000000CleanupOrphanMenus(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		const orphanMenuID = 471

		if err := tx.Exec(
			"DELETE FROM sys_role_menu WHERE menu_id = ?",
			orphanMenuID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			"DELETE FROM sys_menu_api_rule WHERE sys_menu_menu_id = ?",
			orphanMenuID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			"DELETE FROM sys_menu WHERE menu_id = ?",
			orphanMenuID,
		).Error; err != nil {
			return err
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
