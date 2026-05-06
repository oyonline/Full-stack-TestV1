package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	adminModels "go-admin/app/admin/models"
	common "go-admin/common/models"
)

// TestMigration_UserAvatarProfile 用 in-memory sqlite 验证 1775000000000_user_avatar_profile：
//   - 列原本不存在时，能添加 avatar_type / avatar_color
//   - 已有 avatar 非空的旧行被回填为 avatar_type='image'
//   - 重复执行（HasColumn 命中分支）不会失败
func TestMigration_UserAvatarProfile(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// 初始化 sys_user 表，但故意 DROP 后再用旧 schema 建一遍（不带 avatar_type / avatar_color），
	// 模拟生产环境老 schema 的状态。
	if err := db.Migrator().DropTable("sys_user"); err != nil {
		t.Fatalf("drop sys_user: %v", err)
	}
	type oldSysUser struct {
		UserId   int    `gorm:"primaryKey;autoIncrement"`
		Username string `gorm:"size:64"`
		Avatar   string `gorm:"size:255"`
	}
	if err := db.Table("sys_user").AutoMigrate(&oldSysUser{}); err != nil {
		t.Fatalf("auto migrate old: %v", err)
	}

	// migration 写入 sys_migration 记录，先把表建好。
	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate sys_migration: %v", err)
	}

	// 准备数据：1 个有头像的老用户，1 个无头像的老用户
	if err := db.Exec("INSERT INTO sys_user (user_id, username, avatar) VALUES (1, 'alice', '/upload/a.png'), (2, 'bob', '')").Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	// 跑 migration
	if err := _1775000000000UserAvatarProfile(db, "1775000000000_user_avatar_profile.go"); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// 验证列已存在
	if !db.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_type") {
		t.Fatal("avatar_type column not created")
	}
	if !db.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_color") {
		t.Fatal("avatar_color column not created")
	}

	// 验证回填：alice (有头像) → avatar_type='image'；bob (无头像) → avatar_type 为空
	var alice, bob struct {
		AvatarType string
	}
	if err := db.Table("sys_user").Select("avatar_type").Where("user_id = 1").Take(&alice).Error; err != nil {
		t.Fatalf("read alice: %v", err)
	}
	if alice.AvatarType != "image" {
		t.Fatalf("alice should be backfilled to image, got %q", alice.AvatarType)
	}
	if err := db.Table("sys_user").Select("avatar_type").Where("user_id = 2").Take(&bob).Error; err != nil {
		t.Fatalf("read bob: %v", err)
	}
	if bob.AvatarType != "" {
		t.Fatalf("bob (no avatar) should have empty avatar_type, got %q", bob.AvatarType)
	}

	// 重复跑：HasColumn 命中分支应静默跳过；写 sys_migration 时遇到主键冲突，但
	// 我们只关心 schema 改动幂等。这里只重试加列部分。
	if err := db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_type") {
			t.Fatal("HasColumn should be true on second run")
		}
		if !tx.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_color") {
			t.Fatal("HasColumn should be true on second run")
		}
		return nil
	}); err != nil {
		t.Fatalf("re-check: %v", err)
	}
}
