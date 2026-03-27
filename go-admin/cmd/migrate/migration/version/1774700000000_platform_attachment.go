package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"

	"gorm.io/gorm"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1774700000000PlatformAttachment)
}

func _1774700000000PlatformAttachment(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(new(models.AttachmentFile)); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
