package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_SkuModule 用 in-memory sqlite 验证 1779000000000_sku_module：
//   - AutoMigrate 4 张业务表（sku_category / sku_brand / spu / sku）
//   - INSERT sys_menu：'产品中心' 根 + 4 子菜单 + 16 按钮
//   - INSERT sys_api：19 行（SPU 6 + SKU 5 + Category 4 + Brand 4）
//   - INSERT sys_menu_api_rule：19 行（一一对应）
//   - 重跑：已是目标态，FirstOrCreate / 既存检测命中 0 新增；
//     sys_migration 主键冲突需要先清掉记录。
func TestMigration_SkuModule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	type sysMenuApiRule struct {
		SysMenuMenuId int `gorm:"primaryKey;column:sys_menu_menu_id"`
		SysApiId      int `gorm:"primaryKey;column:sys_api_id"`
	}
	if err := db.AutoMigrate(&models.SysMenu{}, &models.SysApi{}); err != nil {
		t.Fatalf("auto migrate sys_menu/sys_api: %v", err)
	}
	if err := db.Table("sys_menu_api_rule").AutoMigrate(&sysMenuApiRule{}); err != nil {
		t.Fatalf("auto migrate sys_menu_api_rule: %v", err)
	}
	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate sys_migration: %v", err)
	}

	const ver = "1779000000000_sku_module.go"

	if err := _1779000000000SkuModule(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// 4 张业务表存在
	for _, name := range []string{"sku_category", "sku_brand", "spu", "sku"} {
		if !db.Migrator().HasTable(name) {
			t.Fatalf("table %s should exist after migration", name)
		}
	}

	// '产品中心' 根菜单（M, parent_id=0）存在且唯一
	var rootCount int64
	if err := db.Table("sys_menu").
		Where("menu_name = ? AND parent_id = ? AND menu_type = ?", "ProductCenter", 0, "M").
		Count(&rootCount).Error; err != nil {
		t.Fatalf("count ProductCenter root: %v", err)
	}
	if rootCount != 1 {
		t.Fatalf("ProductCenter root must be exactly 1, got %d", rootCount)
	}

	// 4 个 C 子菜单（list 权限挂在 C 上）
	listPerms := []string{"admin:spu:list", "admin:sku:list", "admin:category:list", "admin:brand:list"}
	for _, p := range listPerms {
		var c int64
		if err := db.Table("sys_menu").
			Where("permission = ? AND menu_type = ?", p, "C").
			Count(&c).Error; err != nil {
			t.Fatalf("count menu permission=%s: %v", p, err)
		}
		if c != 1 {
			t.Fatalf("menu permission=%s should be 1, got %d", p, c)
		}
	}

	// 4 套按钮（add/edit/remove/query）= 16 行 F 类型
	prefixes := []string{"admin:spu:", "admin:sku:", "admin:category:", "admin:brand:"}
	suffixes := []string{"add", "edit", "remove", "query"}
	for _, prefix := range prefixes {
		for _, suffix := range suffixes {
			perm := prefix + suffix
			var c int64
			if err := db.Table("sys_menu").
				Where("permission = ? AND menu_type = ?", perm, "F").
				Count(&c).Error; err != nil {
				t.Fatalf("count button permission=%s: %v", perm, err)
			}
			if c != 1 {
				t.Fatalf("button permission=%s should be 1, got %d", perm, c)
			}
		}
	}

	// SQL 验证：SELECT FROM sys_menu WHERE permission LIKE 'admin:spu:%' 等返回预期
	for _, prefix := range []string{"admin:spu:%", "admin:sku:%", "admin:category:%", "admin:brand:%"} {
		var c int64
		if err := db.Table("sys_menu").
			Where("permission LIKE ?", prefix).
			Count(&c).Error; err != nil {
			t.Fatalf("count permission LIKE %s: %v", prefix, err)
		}
		// list (C) + 4 buttons (F) = 5 per entity
		if c != 5 {
			t.Fatalf("permission LIKE %s should be 5, got %d", prefix, c)
		}
	}

	// sys_api 19 行（按 path 前缀过滤 SKU 模块）
	var apiCount int64
	if err := db.Table("sys_api").
		Where("path LIKE ? OR path LIKE ? OR path LIKE ? OR path LIKE ?",
			"/api/v1/spu%", "/api/v1/sku%", "/api/v1/sku-category%", "/api/v1/sku-brand%").
		Count(&apiCount).Error; err != nil {
		t.Fatalf("count sku-module sys_api: %v", err)
	}
	if apiCount != 19 {
		t.Fatalf("sku-module sys_api should be 19, got %d", apiCount)
	}

	// sys_menu_api_rule 桥接 19 行（每个 api 都被桥接一次）
	var bridgeCount int64
	if err := db.Table("sys_menu_api_rule").
		Joins("JOIN sys_api ON sys_api.id = sys_menu_api_rule.sys_api_id").
		Where("sys_api.path LIKE ? OR sys_api.path LIKE ? OR sys_api.path LIKE ? OR sys_api.path LIKE ?",
			"/api/v1/spu%", "/api/v1/sku%", "/api/v1/sku-category%", "/api/v1/sku-brand%").
		Count(&bridgeCount).Error; err != nil {
		t.Fatalf("count sku-module sys_menu_api_rule: %v", err)
	}
	if bridgeCount != 19 {
		t.Fatalf("sku-module sys_menu_api_rule should be 19, got %d", bridgeCount)
	}

	// 重跑：已是目标态，所有 FirstOrCreate / count-then-insert 命中 0 新增；
	// sys_migration 主键冲突需要先清掉记录。
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", ver).Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1779000000000SkuModule(db, ver); err != nil {
		t.Fatalf("migration second run: %v", err)
	}

	// 重跑后行数应保持不变
	var rerunMenu, rerunApi, rerunBridge int64
	if err := db.Table("sys_menu").Where("permission LIKE 'admin:spu:%' OR permission LIKE 'admin:sku:%' OR permission LIKE 'admin:category:%' OR permission LIKE 'admin:brand:%'").Count(&rerunMenu).Error; err != nil {
		t.Fatalf("rerun count menus: %v", err)
	}
	if rerunMenu != 20 { // 5 per entity * 4 entities = 20
		t.Fatalf("rerun menus should be 20, got %d", rerunMenu)
	}
	if err := db.Table("sys_api").Where("path LIKE '/api/v1/spu%' OR path LIKE '/api/v1/sku%' OR path LIKE '/api/v1/sku-category%' OR path LIKE '/api/v1/sku-brand%'").Count(&rerunApi).Error; err != nil {
		t.Fatalf("rerun count apis: %v", err)
	}
	if rerunApi != 19 {
		t.Fatalf("rerun apis should be 19, got %d", rerunApi)
	}
	if err := db.Table("sys_menu_api_rule").
		Joins("JOIN sys_api ON sys_api.id = sys_menu_api_rule.sys_api_id").
		Where("sys_api.path LIKE '/api/v1/spu%' OR sys_api.path LIKE '/api/v1/sku%' OR sys_api.path LIKE '/api/v1/sku-category%' OR sys_api.path LIKE '/api/v1/sku-brand%'").
		Count(&rerunBridge).Error; err != nil {
		t.Fatalf("rerun count bridges: %v", err)
	}
	if rerunBridge != 19 {
		t.Fatalf("rerun bridges should be 19, got %d", rerunBridge)
	}
}
