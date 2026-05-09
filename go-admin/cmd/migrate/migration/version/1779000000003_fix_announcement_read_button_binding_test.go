package version

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"
)

// TestMigration_FixAnnouncementReadButtonBinding 用 in-memory sqlite 验证修正逻辑：
//   - fresh install 后（由 1778160000000 建立的错误桥接），read 按钮正确桥接到 MarkRead API
//   - 旧的错误桥接（parentMenu → MarkRead）被删除
//   - 正确桥接（readButton → MarkRead）存在
//   - 重跑幂等：行数保持不变
func TestMigration_FixAnnouncementReadButtonBinding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:readbtnfix?mode=memory&_pragma=foreign_keys(0)"), &gorm.Config{})
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

	// 建父菜单（C 类，admin:announcement:list）
	parentMenu := models.SysMenu{
		MenuName: "AnnouncementManage", Title: "公告管理",
		MenuType: "C", Permission: "admin:announcement:list",
		ParentId: 2, Path: "/admin/sys-announcement",
		Component: "admin/sys-announcement/index",
	}
	if err := db.Create(&parentMenu).Error; err != nil {
		t.Fatalf("seed parentMenu: %v", err)
	}

	// 建 read 按钮（F 类，admin:announcement:read）
	readButton := models.SysMenu{
		Title:      "标记公告已读",
		MenuType:   "F",
		Permission: "admin:announcement:read",
		ParentId:   parentMenu.MenuId,
	}
	if err := db.Create(&readButton).Error; err != nil {
		t.Fatalf("seed readButton: %v", err)
	}

	// 建 MarkRead API
	markReadApi := models.SysApi{
		Handle: "go-admin/app/admin/apis.Announcement.MarkRead-fm",
		Title:  "公告标记已读",
		Path:   "/api/v1/announcement/:id/read",
		Type:   "BUS",
		Action: "POST",
	}
	if err := db.Create(&markReadApi).Error; err != nil {
		t.Fatalf("seed markReadApi: %v", err)
	}

	// 模拟原始迁移的错误桥接：parentMenu → MarkRead
	if err := db.Exec(
		"INSERT INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
		parentMenu.MenuId, markReadApi.Id,
	).Error; err != nil {
		t.Fatalf("seed wrong bridge: %v", err)
	}

	const ver = "1779000000003_fix_announcement_read_button_binding.go"
	if err := _1779000000003FixAnnouncementReadButtonBinding(db, ver); err != nil {
		t.Fatalf("migration run: %v", err)
	}

	// 错误桥接应已删除
	var wrongCount int64
	if err := db.Table("sys_menu_api_rule").
		Where("sys_menu_menu_id = ? AND sys_api_id = ?", parentMenu.MenuId, markReadApi.Id).
		Count(&wrongCount).Error; err != nil {
		t.Fatalf("count wrong bridge: %v", err)
	}
	if wrongCount != 0 {
		t.Fatalf("wrong bridge (parentMenu→MarkRead) should be deleted, got %d rows", wrongCount)
	}

	// 正确桥接应存在
	var correctCount int64
	if err := db.Table("sys_menu_api_rule").
		Where("sys_menu_menu_id = ? AND sys_api_id = ?", readButton.MenuId, markReadApi.Id).
		Count(&correctCount).Error; err != nil {
		t.Fatalf("count correct bridge: %v", err)
	}
	if correctCount != 1 {
		t.Fatalf("correct bridge (readButton→MarkRead) should be 1, got %d rows", correctCount)
	}

	// 重跑：清掉 sys_migration 记录后再跑，行数保持不变
	if err := db.Exec("DELETE FROM sys_migration WHERE version = ?", ver).Error; err != nil {
		t.Fatalf("clear sys_migration: %v", err)
	}
	if err := _1779000000003FixAnnouncementReadButtonBinding(db, ver); err != nil {
		t.Fatalf("migration rerun: %v", err)
	}

	// 重跑后错误桥接仍为 0
	var wrongCount2 int64
	if err := db.Table("sys_menu_api_rule").
		Where("sys_menu_menu_id = ? AND sys_api_id = ?", parentMenu.MenuId, markReadApi.Id).
		Count(&wrongCount2).Error; err != nil {
		t.Fatalf("count wrong bridge after rerun: %v", err)
	}
	if wrongCount2 != 0 {
		t.Fatalf("wrong bridge should remain absent after rerun, got %d rows", wrongCount2)
	}

	// 重跑后正确桥接仍为 1（无重复）
	var correctCount2 int64
	if err := db.Table("sys_menu_api_rule").
		Where("sys_menu_menu_id = ? AND sys_api_id = ?", readButton.MenuId, markReadApi.Id).
		Count(&correctCount2).Error; err != nil {
		t.Fatalf("count correct bridge after rerun: %v", err)
	}
	if correctCount2 != 1 {
		t.Fatalf("correct bridge should remain exactly 1 after rerun, got %d rows", correctCount2)
	}
}

// TestMigration_FixAnnouncementReadButtonBinding_MissingData 验证原始迁移未跑时跳过不报错
func TestMigration_FixAnnouncementReadButtonBinding_MissingData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	type sysMenuApiRule struct {
		SysMenuMenuId int `gorm:"primaryKey;column:sys_menu_menu_id"`
		SysApiId      int `gorm:"primaryKey;column:sys_api_id"`
	}
	if err := db.AutoMigrate(&models.SysMenu{}, &models.SysApi{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Table("sys_menu_api_rule").AutoMigrate(&sysMenuApiRule{}); err != nil {
		t.Fatalf("auto migrate rule: %v", err)
	}
	if err := db.AutoMigrate(&common.Migration{}); err != nil {
		t.Fatalf("auto migrate migration: %v", err)
	}

	// 不插任何数据，模拟未跑过 1778160000000
	const ver = "1779000000003_fix_announcement_read_button_binding.go"
	if err := _1779000000003FixAnnouncementReadButtonBinding(db, ver); err != nil {
		t.Fatalf("migration should not error on empty DB: %v", err)
	}
}
