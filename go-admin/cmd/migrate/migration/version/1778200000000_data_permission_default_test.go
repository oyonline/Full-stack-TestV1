package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	common "go-admin/common/models"
)

// TestMigration_DataPermissionDefault 用 in-memory sqlite 验证
// 1778200000000_data_permission_default：
//   - admin（role_id=1）原本 data_scope=''，迁移后变成 '1'
//   - 历史角色 data_scope IS NULL → 迁移后变成 '1'
//   - 历史角色 data_scope='' → 迁移后变成 '1'
//   - 已填值的角色（'2' / '3' / '5' …）保持原值不动
//   - 重复执行不报错（命中 0 行 + sys_migration 版本号清理后再写）
func TestMigration_DataPermissionDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	type sysRole struct {
		RoleId    int     `gorm:"primaryKey"`
		RoleName  string  `gorm:"size:128"`
		DataScope *string `gorm:"size:128"`
	}
	if err := db.Table("sys_role").AutoMigrate(&sysRole{}); err != nil {
		t.Fatalf("auto migrate sys_role: %v", err)
	}
	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate sys_migration: %v", err)
	}

	emptyStr := ""
	scope2 := "2"
	scope3 := "3"
	scope5 := "5"
	seedRows := []sysRole{
		{RoleId: 1, RoleName: "系统管理员", DataScope: &emptyStr},
		{RoleId: 2, RoleName: "历史角色-NULL", DataScope: nil},
		{RoleId: 3, RoleName: "历史角色-空串", DataScope: &emptyStr},
		{RoleId: 4, RoleName: "本部门", DataScope: &scope2},
		{RoleId: 5, RoleName: "本部门及下级", DataScope: &scope3},
		{RoleId: 6, RoleName: "自定义", DataScope: &scope5},
	}
	if err := db.Table("sys_role").Create(&seedRows).Error; err != nil {
		t.Fatalf("seed sys_role: %v", err)
	}

	if err := _1778200000000DataPermissionDefault(db, "1778200000000_data_permission_default.go"); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	want := map[int]string{
		1: "1",
		2: "1",
		3: "1",
		4: "2",
		5: "3",
		6: "5",
	}
	for roleId, expected := range want {
		var got sysRole
		if err := db.Table("sys_role").Where("role_id = ?", roleId).Take(&got).Error; err != nil {
			t.Fatalf("query sys_role role_id=%d: %v", roleId, err)
		}
		if got.DataScope == nil {
			t.Fatalf("sys_role role_id=%d data_scope should be %q, got NULL", roleId, expected)
		}
		if *got.DataScope != expected {
			t.Fatalf("sys_role role_id=%d data_scope want %q, got %q", roleId, expected, *got.DataScope)
		}
	}

	// 验收 #2：admin (role_id=1) data_scope='1'
	var adminScope sysRole
	if err := db.Table("sys_role").Where("role_id = ?", 1).Take(&adminScope).Error; err != nil {
		t.Fatalf("re-query admin: %v", err)
	}
	if adminScope.DataScope == nil || *adminScope.DataScope != "1" {
		t.Fatalf("admin data_scope must be '1' after migration")
	}

	// 验收 #1：全表无 NULL / 空 data_scope
	var nullOrEmpty int64
	if err := db.Table("sys_role").
		Where("data_scope IS NULL OR data_scope = ?", "").
		Count(&nullOrEmpty).Error; err != nil {
		t.Fatalf("count null/empty data_scope: %v", err)
	}
	if nullOrEmpty != 0 {
		t.Fatalf("after migration sys_role must have no NULL/empty data_scope, got %d", nullOrEmpty)
	}

	// 重跑：已是目标态，UPDATE 命中 0 行不报错；sys_migration 主键冲突需要先清掉。
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", "1778200000000_data_permission_default.go").Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1778200000000DataPermissionDefault(db, "1778200000000_data_permission_default.go"); err != nil {
		t.Fatalf("migration second run: %v", err)
	}
}
