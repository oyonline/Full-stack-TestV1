package version

import (
	"runtime"

	adminModels "go-admin/app/admin/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1775000000000UserAvatarProfile)
}

// _1775000000000UserAvatarProfile 把头像扩展字段加到 sys_user 表，并把已有 avatar 数据回填为 image 模式。
//
// 加列：
//   - avatar_type  varchar(16)：image / letter / "" — 决定前端渲染头像图片还是字母色块。
//   - avatar_color varchar(16)：hex 颜色，例如 #1D4ED8 — letter 模式下的背景色。
//
// 数据回填：
//   - 已有 avatar 字段非空的旧用户写成 avatar_type='image'，对齐 docs/sql 参考脚本，
//     避免老用户首屏头像突然回退成字母色块。
func _1775000000000UserAvatarProfile(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_type") {
			if err := tx.Migrator().AddColumn(&adminModels.SysUser{}, "AvatarType"); err != nil {
				return err
			}
		}
		if !tx.Migrator().HasColumn(&adminModels.SysUser{}, "avatar_color") {
			if err := tx.Migrator().AddColumn(&adminModels.SysUser{}, "AvatarColor"); err != nil {
				return err
			}
		}
		// 兼容老数据：曾经上传过头像图片的用户回填为 image 模式。
		if err := tx.Exec(
			"UPDATE sys_user SET avatar_type = ? WHERE avatar <> '' AND (avatar_type IS NULL OR avatar_type = '')",
			"image",
		).Error; err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
