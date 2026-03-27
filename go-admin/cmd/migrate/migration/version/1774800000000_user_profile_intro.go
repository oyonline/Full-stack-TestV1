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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1774800000000UserProfileIntro)
}

func _1774800000000UserProfileIntro(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasColumn(&adminModels.SysUser{}, "introduction") {
			if err := tx.Migrator().AddColumn(&adminModels.SysUser{}, "Introduction"); err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
