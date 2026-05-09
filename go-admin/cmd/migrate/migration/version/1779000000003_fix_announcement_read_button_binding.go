package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000003FixAnnouncementReadButtonBinding)
}

// _1779000000003FixAnnouncementReadButtonBinding 修正 admin:announcement:read 按钮挂载位置。
//
// 1778160000000_announcement.go:119 把 MarkRead API (POST /api/v1/announcement/:id/read)
// 桥接到了父菜单（list 权限），导致 admin:announcement:read 按钮权限实际无效，
// announcement_read_log 无法正确写入。
//
// 本迁移：
//  1. 删除错误桥接行（parentMenu.MenuId → MarkRead API）
//  2. 插入正确桥接行（readButton.MenuId → MarkRead API），幂等
func _1779000000003FixAnnouncementReadButtonBinding(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 找父菜单 ID（C 类，admin:announcement:list）
		var parentMenuId int
		if err := tx.Table("sys_menu").
			Select("menu_id").
			Where("permission = ? AND menu_type = ?", "admin:announcement:list", "C").
			Scan(&parentMenuId).Error; err != nil {
			return err
		}

		// 找 read 按钮 ID（F 类，admin:announcement:read）
		var readButtonId int
		if err := tx.Table("sys_menu").
			Select("menu_id").
			Where("permission = ? AND menu_type = ?", "admin:announcement:read", "F").
			Scan(&readButtonId).Error; err != nil {
			return err
		}

		// 找 MarkRead API ID（POST /api/v1/announcement/:id/read）
		var markReadApiId int
		if err := tx.Table("sys_api").
			Select("id").
			Where("path = ? AND action = ?", "/api/v1/announcement/:id/read", "POST").
			Scan(&markReadApiId).Error; err != nil {
			return err
		}

		// 任意 ID 为 0 说明原始迁移未跑或 DB 状态异常，跳过
		if parentMenuId == 0 || readButtonId == 0 || markReadApiId == 0 {
			return tx.Create(&common.Migration{Version: version}).Error
		}

		// 1) 删除错误桥接行（parentMenu → MarkRead），幂等
		if err := tx.Exec(
			"DELETE FROM sys_menu_api_rule WHERE sys_menu_menu_id = ? AND sys_api_id = ?",
			parentMenuId, markReadApiId,
		).Error; err != nil {
			return err
		}

		// 2) 插入正确桥接行（readButton → MarkRead），先检查避免 SQL 方言差异
		var n int64
		if err := tx.Table("sys_menu_api_rule").
			Where("sys_menu_menu_id = ? AND sys_api_id = ?", readButtonId, markReadApiId).
			Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := tx.Exec(
				"INSERT INTO sys_menu_api_rule (sys_menu_menu_id, sys_api_id) VALUES (?, ?)",
				readButtonId, markReadApiId,
			).Error; err != nil {
				return err
			}
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
