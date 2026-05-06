package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1774900000000DropFinanceSubsystem)
}

// _1774900000000DropFinanceSubsystem 清理 finance 子系统在 MySQL 中的残留：
// 1) DROP TABLE 11 张 finance 业务表（commit f5b3312 已删除对应 Go 代码）
// 2) DELETE sys_api 中 finance 路由前缀的接口注册条目
// 3) DELETE sys_menu 中 finance 模块的菜单（path / permission / component 命中）
// 4) 级联清理 sys_role_menu、sys_menu_api_rule
// 5) DELETE sys_casbin_rule 中 V1 命中 finance API 路径的策略
// 6) DELETE module_registry 中 finance-budget 模块条目（1774600000000 创建）
//
// 设计说明：
//   - sys_menu / sys_api 真相源是数据库（AGENTS.md），无 seed 文件维护，故按
//     路由/权限前缀匹配；前缀来自被删除的 router/*.go 文件（git show f5b3312）。
//   - sys_role_menu / sys_menu_api_rule 用子查询级联，避免依赖具体 menu_id。
//   - 全部使用 IF EXISTS / 条件 DELETE，幂等可重跑。
func _1774900000000DropFinanceSubsystem(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		financeTables := []string{
			"cost_center_info",
			"cost_center_info_change",
			"cost_center_related_customer",
			"cost_budget_version",
			"cost_budget_version_detail",
			"budget_fee_category",
			"budget_fee_category_details",
			"allocation_rule_settings",
			"allocation_rule_settings_dept",
			"fee_request_log",
			"fee_request",
		}

		// 1) Drop business tables
		for _, t := range financeTables {
			if err := tx.Migrator().DropTable(t); err != nil {
				return err
			}
		}

		// 2) sys_api：路径前缀来自被删除的 router 文件
		apiPathPrefixes := []string{
			"/api/v1/cost-center-info",
			"/api/v1/cost-center-info-change",
			"/api/v1/cost-center-related-customer",
			"/api/v1/cost-budget-version",
			"/api/v1/cost-budget-version-detail",
			"/api/v1/budget-fee-category",
			"/api/v1/budget-fee-category-details",
			"/api/v1/allocation-rule-settings",
			"/api/v1/allocation-rule-settings-dept",
			"/api/v1/fee-request",
		}

		// 5) Casbin：V1 是策略中的 path，命中同样的 finance API 路径前缀
		for _, p := range apiPathPrefixes {
			like := p + "%"
			if err := tx.Exec("DELETE FROM sys_casbin_rule WHERE ptype = ? AND v1 LIKE ?", "p", like).Error; err != nil {
				return err
			}
		}

		// 4a) sys_menu_api_rule：先按 sys_api 子查询删 m2m 关系
		for _, p := range apiPathPrefixes {
			like := p + "%"
			if err := tx.Exec(
				"DELETE FROM sys_menu_api_rule WHERE sys_api_id IN (SELECT id FROM sys_api WHERE path LIKE ?)",
				like,
			).Error; err != nil {
				return err
			}
		}

		// 2-cont) 删 sys_api 本体
		for _, p := range apiPathPrefixes {
			like := p + "%"
			if err := tx.Exec("DELETE FROM sys_api WHERE path LIKE ?", like).Error; err != nil {
				return err
			}
		}

		// 3 + 4b) sys_menu：按 path / permission / component 命中 finance
		// path/permission/component 的前缀基于前端路由约定与 module_registry 中
		// finance-budget 模块的 RouteBase=/finance/budget、PermissionHint=finance-budget
		menuMatchers := []struct {
			column string
			like   string
		}{
			{"path", "/finance/%"},
			{"path", "/finance"},
			{"permission", "finance:%"},
			{"permission", "finance-budget:%"},
			{"permission", "cost-center%"},
			{"permission", "cost-budget%"},
			{"permission", "budget-fee%"},
			{"permission", "allocation-rule%"},
			{"permission", "fee-request%"},
			{"component", "%finance/%"},
			{"component", "%cost-center%"},
			{"component", "%cost-budget%"},
			{"component", "%budget-fee%"},
			{"component", "%allocation-rule%"},
			{"component", "%fee-request%"},
		}

		for _, m := range menuMatchers {
			// 先删 sys_role_menu 关系
			if err := tx.Exec(
				"DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_id FROM sys_menu WHERE "+m.column+" LIKE ?)",
				m.like,
			).Error; err != nil {
				return err
			}
			// 再删 sys_menu_api_rule 中以这些菜单为左侧的关系
			if err := tx.Exec(
				"DELETE FROM sys_menu_api_rule WHERE sys_menu_menu_id IN (SELECT menu_id FROM sys_menu WHERE "+m.column+" LIKE ?)",
				m.like,
			).Error; err != nil {
				return err
			}
			// 最后删菜单本体
			if err := tx.Exec("DELETE FROM sys_menu WHERE "+m.column+" LIKE ?", m.like).Error; err != nil {
				return err
			}
		}

		// 6) module_registry：删 finance-budget 模块
		if err := tx.Exec("DELETE FROM module_registry WHERE module_key = ?", "finance-budget").Error; err != nil {
			return err
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
