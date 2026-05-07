package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	common "go-admin/common/models"
)

// TestMigration_CleanupOrphanMenus 用 in-memory sqlite 验证 1775400000000_cleanup_orphan_menus：
//   - 删 menu_id=471（含 sys_role_menu / sys_menu_api_rule 配套清理）
//   - 不动同父级（459 Schedule）下兄弟 460 ScheduleManage 与其子 461-464
//   - 不动 spec 原表里被误判的 6 个真实页菜单（262/61/211/460/269/537）
//   - 重复执行（已是目标态）不会失败
func TestMigration_CleanupOrphanMenus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(0)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	type sysMenu struct {
		MenuId    int    `gorm:"primaryKey"`
		MenuName  string `gorm:"size:128"`
		Title     string `gorm:"size:128"`
		Path      string `gorm:"size:128"`
		Paths     string `gorm:"size:128"`
		Component string `gorm:"size:255"`
		ParentId  int
	}
	type sysRoleMenu struct {
		RoleId int `gorm:"primaryKey"`
		MenuId int `gorm:"primaryKey;column:menu_id"`
	}
	type sysMenuApiRule struct {
		SysMenuMenuId int `gorm:"primaryKey;column:sys_menu_menu_id"`
		SysApiId      int `gorm:"primaryKey;column:sys_api_id"`
	}

	if err := db.Table("sys_menu").AutoMigrate(&sysMenu{}); err != nil {
		t.Fatalf("auto migrate sys_menu: %v", err)
	}
	if err := db.Table("sys_role_menu").AutoMigrate(&sysRoleMenu{}); err != nil {
		t.Fatalf("auto migrate sys_role_menu: %v", err)
	}
	if err := db.Table("sys_menu_api_rule").AutoMigrate(&sysMenuApiRule{}); err != nil {
		t.Fatalf("auto migrate sys_menu_api_rule: %v", err)
	}
	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate sys_migration: %v", err)
	}

	// seed sys_menu：471（要删）+ 父 459 + 兄弟 460 + 460 的 F 子按钮 461-464 +
	// spec 原列表中被误判的 6 个真实页（验证不被误伤）。
	seedRows := []sysMenu{
		{MenuId: 459, MenuName: "Schedule", Title: "定时任务", Path: "/schedule", Paths: "/0/459", Component: "Layout", ParentId: 0},
		{MenuId: 460, MenuName: "ScheduleManage", Title: "Schedule", Path: "/admin/sys-job", Paths: "/0/459/460", Component: "/admin/sys-job/index", ParentId: 459},
		{MenuId: 461, Title: "分页获取定时任务", Paths: "/0/459/460/461", ParentId: 460},
		{MenuId: 462, Title: "创建定时任务", Paths: "/0/459/460/462", ParentId: 460},
		{MenuId: 463, Title: "修改定时任务", Paths: "/0/459/460/463", ParentId: 460},
		{MenuId: 464, Title: "删除定时任务", Paths: "/0/459/460/464", ParentId: 460},
		{MenuId: 471, MenuName: "JobLog", Title: "日志", Path: "/schedule/log", Paths: "/0/459/471", Component: "/schedule/log", ParentId: 459},
		// 被原 spec 误判，必须保留
		{MenuId: 61, MenuName: "Swagger", Path: "/admin/sys-api", Paths: "/0/60/61", Component: "/admin/sys-api/index", ParentId: 60},
		{MenuId: 211, MenuName: "Log", Path: "/log", Paths: "/0/2/211", Component: "RouteView", ParentId: 2},
		{MenuId: 262, MenuName: "EditTable", Path: "/dev-tools/editTable", Paths: "/0/60/262", Component: "/dev-tools/gen/edit", ParentId: 60},
		{MenuId: 269, MenuName: "ServerMonitor", Path: "/admin/sys-server-monitor", Paths: "/0/60/269", Component: "/admin/sys-server-monitor/index", ParentId: 537},
		{MenuId: 537, MenuName: "SysTools", Path: "/sys-tools", Paths: "", Component: "Layout", ParentId: 0},
	}
	if err := db.Table("sys_menu").Create(&seedRows).Error; err != nil {
		t.Fatalf("seed sys_menu: %v", err)
	}

	// seed sys_role_menu：admin (role_id=1) 持有 471 + 460 + 462 + 269（保留组）
	if err := db.Table("sys_role_menu").Create(&[]sysRoleMenu{
		{RoleId: 1, MenuId: 471},
		{RoleId: 1, MenuId: 460},
		{RoleId: 1, MenuId: 462},
		{RoleId: 1, MenuId: 269},
	}).Error; err != nil {
		t.Fatalf("seed sys_role_menu: %v", err)
	}

	// seed sys_menu_api_rule：471 + 460（保留组）
	if err := db.Table("sys_menu_api_rule").Create(&[]sysMenuApiRule{
		{SysMenuMenuId: 471, SysApiId: 100},
		{SysMenuMenuId: 460, SysApiId: 100},
	}).Error; err != nil {
		t.Fatalf("seed sys_menu_api_rule: %v", err)
	}

	if err := _1775400000000CleanupOrphanMenus(db, "1775400000000_cleanup_orphan_menus.go"); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// 471 全清
	var menuCount int64
	if err := db.Table("sys_menu").Where("menu_id = ?", 471).Count(&menuCount).Error; err != nil {
		t.Fatalf("count sys_menu 471: %v", err)
	}
	if menuCount != 0 {
		t.Fatalf("sys_menu menu_id=471 should be deleted, got count=%d", menuCount)
	}
	var rmCount int64
	if err := db.Table("sys_role_menu").Where("menu_id = ?", 471).Count(&rmCount).Error; err != nil {
		t.Fatalf("count sys_role_menu 471: %v", err)
	}
	if rmCount != 0 {
		t.Fatalf("sys_role_menu menu_id=471 should be deleted, got count=%d", rmCount)
	}
	var arCount int64
	if err := db.Table("sys_menu_api_rule").Where("sys_menu_menu_id = ?", 471).Count(&arCount).Error; err != nil {
		t.Fatalf("count sys_menu_api_rule 471: %v", err)
	}
	if arCount != 0 {
		t.Fatalf("sys_menu_api_rule sys_menu_menu_id=471 should be deleted, got count=%d", arCount)
	}

	// 兄弟 460 与其 F 子按钮 461-464 + 父级 459 全部保留
	preserved := []int{459, 460, 461, 462, 463, 464}
	for _, id := range preserved {
		var c int64
		if err := db.Table("sys_menu").Where("menu_id = ?", id).Count(&c).Error; err != nil {
			t.Fatalf("count sys_menu %d: %v", id, err)
		}
		if c != 1 {
			t.Fatalf("sys_menu menu_id=%d should be preserved, got count=%d", id, c)
		}
	}

	// 原 spec 误判的 6 个 menu_id 全部保留
	misidentified := []int{262, 61, 211, 460, 269, 537}
	for _, id := range misidentified {
		var c int64
		if err := db.Table("sys_menu").Where("menu_id = ?", id).Count(&c).Error; err != nil {
			t.Fatalf("count sys_menu %d: %v", id, err)
		}
		if c != 1 {
			t.Fatalf("sys_menu menu_id=%d (live page misidentified by original spec) should be preserved, got count=%d", id, c)
		}
	}

	// 兄弟 460 的关联保留
	if err := db.Table("sys_role_menu").Where("menu_id = ?", 460).Count(&rmCount).Error; err != nil {
		t.Fatalf("count sys_role_menu 460: %v", err)
	}
	if rmCount != 1 {
		t.Fatalf("sys_role_menu menu_id=460 should be preserved, got count=%d", rmCount)
	}
	if err := db.Table("sys_menu_api_rule").Where("sys_menu_menu_id = ?", 460).Count(&arCount).Error; err != nil {
		t.Fatalf("count sys_menu_api_rule 460: %v", err)
	}
	if arCount != 1 {
		t.Fatalf("sys_menu_api_rule sys_menu_menu_id=460 should be preserved, got count=%d", arCount)
	}

	// 重跑：已是目标态，DELETE 命中 0 行不报错；写 sys_migration 主键冲突需要先清掉记录。
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", "1775400000000_cleanup_orphan_menus.go").Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1775400000000CleanupOrphanMenus(db, "1775400000000_cleanup_orphan_menus.go"); err != nil {
		t.Fatalf("migration second run: %v", err)
	}
}
