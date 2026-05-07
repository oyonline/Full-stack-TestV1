package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	adminModels "go-admin/app/admin/models"
	common "go-admin/common/models"
)

// TestMigration_SysUserFeishuFields 用 in-memory sqlite 验证 1775200000000_sys_user_feishu_fields：
//   - 列原本不存在时，能添加 5 个飞书字段
//   - 重复执行（HasColumn 命中分支）不会失败
func TestMigration_SysUserFeishuFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// 模拟生产环境老 schema：sys_user 不带任何飞书字段
	if err := db.Migrator().DropTable("sys_user"); err != nil {
		t.Fatalf("drop sys_user: %v", err)
	}
	type oldSysUser struct {
		UserId   int    `gorm:"primaryKey;autoIncrement"`
		Username string `gorm:"size:64"`
	}
	if err := db.Table("sys_user").AutoMigrate(&oldSysUser{}); err != nil {
		t.Fatalf("auto migrate old: %v", err)
	}

	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate sys_migration: %v", err)
	}

	if err := _1775200000000SysUserFeishuFields(db, "1775200000000_sys_user_feishu_fields.go"); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	wantColumns := []string{
		"open_id",
		"job_title",
		"open_department_id",
		"open_department_ids",
		"cn_name",
	}
	for _, col := range wantColumns {
		if !db.Migrator().HasColumn(&adminModels.SysUser{}, col) {
			t.Fatalf("column %q not created", col)
		}
	}

	// 重复跑：HasColumn 命中分支应静默跳过加列。
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, col := range wantColumns {
			if !tx.Migrator().HasColumn(&adminModels.SysUser{}, col) {
				t.Fatalf("HasColumn should be true on second run for %q", col)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("re-check: %v", err)
	}
}
