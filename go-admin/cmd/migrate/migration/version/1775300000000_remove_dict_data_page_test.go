package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	common "go-admin/common/models"
)

// TestMigration_RemoveDictDataPage 用 in-memory sqlite 验证 1775300000000_remove_dict_data_page：
//   - 删 menu_id=59、240（含 sys_role_menu / sys_menu_api_rule 配套清理）
//   - 把 menu_id=241/242/243 重挂到 543 下，paths 改为 '/0/2/58/543/{id}'
//   - 不动 menu_id=543、236（已挂 543 的兄弟节点）等无关行
//   - 重复执行（数据已是目标态）不会失败
func TestMigration_RemoveDictDataPage(t *testing.T) {
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

	// seed sys_menu：被影响的字典相关行 + 一个无关兄弟（236）保证不被误伤
	seedRows := []sysMenu{
		{MenuId: 59, MenuName: "SysDictDataManage", Title: "字典数据", Path: "/admin/dict/data/:dictId", Paths: "/0/2/58/59", Component: "/admin/sys-dict-data/index", ParentId: 58},
		{MenuId: 240, Title: "查询数据", Paths: "/0/2/58/59/240", ParentId: 59},
		{MenuId: 241, Title: "新增数据", Paths: "/0/2/58/59/241", ParentId: 59},
		{MenuId: 242, Title: "修改数据", Paths: "/0/2/58/59/242", ParentId: 59},
		{MenuId: 243, Title: "删除数据", Paths: "/0/2/58/59/243", ParentId: 59},
		{MenuId: 543, MenuName: "SysDictTypeManage", Title: "字典类型", Path: "/admin/sys-dict-type", Paths: "/0/2/58/543", Component: "/admin/sys-dict-type/index", ParentId: 58},
		{MenuId: 236, Title: "查询字典", Paths: "/0/2/58/543/236", ParentId: 543},
	}
	if err := db.Table("sys_menu").Create(&seedRows).Error; err != nil {
		t.Fatalf("seed sys_menu: %v", err)
	}

	// seed sys_role_menu：admin 角色（role_id=1）持有要删的菜单和要保留的菜单
	if err := db.Table("sys_role_menu").Create(&[]sysRoleMenu{
		{RoleId: 1, MenuId: 59},
		{RoleId: 1, MenuId: 240},
		{RoleId: 1, MenuId: 241},
		{RoleId: 1, MenuId: 543},
	}).Error; err != nil {
		t.Fatalf("seed sys_role_menu: %v", err)
	}

	// seed sys_menu_api_rule：59、240 配套关联 + 241 和 543 自己的关联（应保留）
	if err := db.Table("sys_menu_api_rule").Create(&[]sysMenuApiRule{
		{SysMenuMenuId: 59, SysApiId: 24},
		{SysMenuMenuId: 240, SysApiId: 24},
		{SysMenuMenuId: 241, SysApiId: 80},
		{SysMenuMenuId: 543, SysApiId: 21},
	}).Error; err != nil {
		t.Fatalf("seed sys_menu_api_rule: %v", err)
	}

	// 跑 migration
	if err := _1775300000000RemoveDictDataPage(db, "1775300000000_remove_dict_data_page.go"); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// 1) menu_id=59、240 被删
	for _, id := range []int{59, 240} {
		var cnt int64
		if err := db.Table("sys_menu").Where("menu_id = ?", id).Count(&cnt).Error; err != nil {
			t.Fatalf("count menu %d: %v", id, err)
		}
		if cnt != 0 {
			t.Fatalf("menu_id=%d should be deleted, got %d rows", id, cnt)
		}
	}

	// 2) 241/242/243 被迁到 543
	for _, id := range []int{241, 242, 243} {
		var row sysMenu
		if err := db.Table("sys_menu").Where("menu_id = ?", id).Take(&row).Error; err != nil {
			t.Fatalf("read menu %d: %v", id, err)
		}
		if row.ParentId != 543 {
			t.Fatalf("menu_id=%d parent_id should be 543, got %d", id, row.ParentId)
		}
		wantPaths := "/0/2/58/543/" + map[int]string{241: "241", 242: "242", 243: "243"}[id]
		if row.Paths != wantPaths {
			t.Fatalf("menu_id=%d paths should be %q, got %q", id, wantPaths, row.Paths)
		}
	}

	// 3) 543 自身和兄弟（236）未被动过
	var keep sysMenu
	if err := db.Table("sys_menu").Where("menu_id = 543").Take(&keep).Error; err != nil {
		t.Fatalf("read menu 543: %v", err)
	}
	if keep.ParentId != 58 || keep.Paths != "/0/2/58/543" {
		t.Fatalf("menu 543 should be untouched, got parent=%d paths=%q", keep.ParentId, keep.Paths)
	}
	var sibling sysMenu
	if err := db.Table("sys_menu").Where("menu_id = 236").Take(&sibling).Error; err != nil {
		t.Fatalf("read menu 236: %v", err)
	}
	if sibling.ParentId != 543 || sibling.Paths != "/0/2/58/543/236" {
		t.Fatalf("menu 236 should be untouched, got parent=%d paths=%q", sibling.ParentId, sibling.Paths)
	}

	// 4) sys_role_menu 中 59、240 配套行被清理；241、543 的关系保留
	for _, id := range []int{59, 240} {
		var cnt int64
		if err := db.Table("sys_role_menu").Where("menu_id = ?", id).Count(&cnt).Error; err != nil {
			t.Fatalf("count role_menu %d: %v", id, err)
		}
		if cnt != 0 {
			t.Fatalf("sys_role_menu menu_id=%d should be deleted, got %d", id, cnt)
		}
	}
	for _, id := range []int{241, 543} {
		var cnt int64
		if err := db.Table("sys_role_menu").Where("menu_id = ?", id).Count(&cnt).Error; err != nil {
			t.Fatalf("count role_menu %d: %v", id, err)
		}
		if cnt == 0 {
			t.Fatalf("sys_role_menu menu_id=%d should remain", id)
		}
	}

	// 5) sys_menu_api_rule 同样清理
	for _, id := range []int{59, 240} {
		var cnt int64
		if err := db.Table("sys_menu_api_rule").Where("sys_menu_menu_id = ?", id).Count(&cnt).Error; err != nil {
			t.Fatalf("count api_rule %d: %v", id, err)
		}
		if cnt != 0 {
			t.Fatalf("sys_menu_api_rule sys_menu_menu_id=%d should be deleted, got %d", id, cnt)
		}
	}
	for _, id := range []int{241, 543} {
		var cnt int64
		if err := db.Table("sys_menu_api_rule").Where("sys_menu_menu_id = ?", id).Count(&cnt).Error; err != nil {
			t.Fatalf("count api_rule %d: %v", id, err)
		}
		if cnt == 0 {
			t.Fatalf("sys_menu_api_rule sys_menu_menu_id=%d should remain", id)
		}
	}

	// 6) 重跑：当前数据已是目标态，DELETE/UPDATE 命中 0 行也不应报错
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_menu WHERE menu_id IN ?", []int{59, 240}).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_menu_api_rule WHERE sys_menu_menu_id IN ?", []int{59, 240}).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_menu WHERE menu_id IN ?", []int{59, 240}).Error; err != nil {
			return err
		}
		return tx.Exec("UPDATE sys_menu SET parent_id = 543, paths = '/0/2/58/543/241' WHERE menu_id = 241").Error
	}); err != nil {
		t.Fatalf("re-run idempotent slice: %v", err)
	}
}
