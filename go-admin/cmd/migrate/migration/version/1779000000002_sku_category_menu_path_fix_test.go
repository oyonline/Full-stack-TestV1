package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_SkuCategoryMenuPathFix 用 in-memory sqlite 验证修正逻辑：
//   - 当 SkuCategoryManage.component 为旧值 'admin/sku-category/index' 时，更新为新值
//   - 当 component 已是新值或被人工改成其它值时，不动
//   - 重跑：sys_migration 主键冲突需要先清掉记录，行数保持稳定
func TestMigration_SkuCategoryMenuPathFix(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:menupathfix?mode=memory&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.SysMenu{}, &common.Migration{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	// 1) 旧值 -> 应被修正
	old := models.SysMenu{
		MenuName: "SkuCategoryManage", Title: "类目管理", Path: "/product/category",
		MenuType: "C", Permission: "admin:category:list",
		ParentId: 1, Component: "admin/sku-category/index",
	}
	if err := db.Create(&old).Error; err != nil {
		t.Fatalf("seed old menu: %v", err)
	}

	// 2) 已被人工改成另一路径 -> 不应被覆盖
	custom := models.SysMenu{
		MenuName: "SkuCategoryManageCustom", Title: "类目管理-自定义",
		Path: "/product/category-custom", MenuType: "C",
		Permission: "admin:category:list:x",
		ParentId:   1, Component: "admin/manual-override/index",
	}
	if err := db.Create(&custom).Error; err != nil {
		t.Fatalf("seed custom menu: %v", err)
	}

	const ver = "1779000000002_sku_category_menu_path_fix.go"
	if err := _1779000000002SkuCategoryMenuPathFix(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	var got models.SysMenu
	if err := db.Where("menu_name = ?", "SkuCategoryManage").First(&got).Error; err != nil {
		t.Fatalf("read updated menu: %v", err)
	}
	if got.Component != "admin/sys-sku-category/index" {
		t.Fatalf("component should be updated, got %q", got.Component)
	}

	var customGot models.SysMenu
	if err := db.Where("menu_name = ?", "SkuCategoryManageCustom").First(&customGot).Error; err != nil {
		t.Fatalf("read custom menu: %v", err)
	}
	if customGot.Component != "admin/manual-override/index" {
		t.Fatalf("manual override component must not be touched, got %q", customGot.Component)
	}

	// 重跑：清掉 sys_migration 记录后再跑，行数保持不变
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", ver).Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1779000000002SkuCategoryMenuPathFix(db, ver); err != nil {
		t.Fatalf("migration rerun: %v", err)
	}

	var menuCount int64
	if err := db.Table("sys_menu").
		Where("menu_name IN ?", []string{"SkuCategoryManage", "SkuCategoryManageCustom"}).
		Count(&menuCount).Error; err != nil {
		t.Fatalf("count menus: %v", err)
	}
	if menuCount != 2 {
		t.Fatalf("expected 2 menu rows after rerun, got %d", menuCount)
	}

	// 第二次跑后 SkuCategoryManage.component 仍为新值
	var got2 models.SysMenu
	if err := db.Where("menu_name = ?", "SkuCategoryManage").First(&got2).Error; err != nil {
		t.Fatalf("read after rerun: %v", err)
	}
	if got2.Component != "admin/sys-sku-category/index" {
		t.Fatalf("component should remain new value after rerun, got %q", got2.Component)
	}
}
