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
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1774600000000PlatformCore)
}

func _1774600000000PlatformCore(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			new(models.ModuleRegistry),
			new(models.WorkflowDefinition),
			new(models.WorkflowDefinitionNode),
			new(models.WorkflowInstance),
			new(models.WorkflowTask),
			new(models.WorkflowActionLog),
			new(models.WorkflowBusinessBinding),
		); err != nil {
			return err
		}

		defaultModules := []models.ModuleRegistry{
			{
				ModuleKey:      "finance-budget",
				ModuleName:     "财务预算管控",
				RouteBase:      "/finance/budget",
				MenuRootCode:   "finance:budget",
				Status:         "2",
				Sort:           10,
				PermissionHint: "finance-budget",
				Remark:         "首个真实业务模块接入样板",
			},
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "module_key"}},
			DoNothing: true,
		}).Create(&defaultModules).Error; err != nil {
			return err
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}
