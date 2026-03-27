package version

import (
	"runtime"

	"go-admin/cmd/migrate/migration"
	"go-admin/cmd/migrate/migration/models"
	common "go-admin/common/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1774502400000SysUserRole)
}

func _1774502400000SysUserRole(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(new(models.SysUserRole)); err != nil {
			return err
		}
		type legacyUserRole struct {
			UserId int `gorm:"column:user_id"`
			RoleId int `gorm:"column:role_id"`
		}
		legacyRows := make([]legacyUserRole, 0)
		if err := tx.Table("sys_user").Select("user_id, role_id").Where("role_id > 0").Find(&legacyRows).Error; err != nil {
			return err
		}
		if len(legacyRows) > 0 {
			records := make([]models.SysUserRole, 0, len(legacyRows))
			for _, row := range legacyRows {
				records = append(records, models.SysUserRole{
					UserId:    row.UserId,
					RoleId:    row.RoleId,
					IsPrimary: true,
				})
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error; err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
