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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1779000000004ModuleRegistryAdminSeed)
}

// _1779000000004ModuleRegistryAdminSeed 向 module_registry 注入 admin 业务模块。
//
// admin 模块是 SPU/SKU 与 AnnouncementAttachmentGC 等功能的宿主模块，
// module_key 为 "admin"，与 workflow/attachment 服务中 EnsureModuleEnabled 的校验值一致。
// 使用 OnConflict DoNothing 保证幂等。
func _1779000000004ModuleRegistryAdminSeed(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		modules := []models.ModuleRegistry{
			{
				ModuleKey:      "admin",
				ModuleName:     "后台管理",
				RouteBase:      "/admin",
				MenuRootCode:   "admin",
				Status:         "2",
				Sort:           1,
				PermissionHint: "admin",
				Remark:         "平台基础业务模块（SPU/SKU/公告等）",
			},
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "module_key"}},
			DoNothing: true,
		}).Create(&modules).Error; err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
